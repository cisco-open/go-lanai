package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
)

type TokenRefresher struct {
	client    Client
	renewer   *api.Renewer
	refreshCh chan error
}

const renewerDescription = "vault apiClient token"

func NewTokenRefresher(client *Client) *TokenRefresher {
	renewer, err := client.GetClientTokenRenewer()
	if err != nil {
		return nil
	}

	return &TokenRefresher{
		client:    *client,
		renewer:   renewer,
		refreshCh: make(chan error),
	}
}

func (r *TokenRefresher) Start(ctx context.Context) {
	// These two go routine exits when the renewer is stopped
	//r.Renew() starts a blocking process to periodically renew the token. Therefore we run it as a go routine
	go r.renewer.Renew()
	//this starts a background process to log the renewal events.
	go r.MonitorRenew(ctx)

	go r.startRefresher(ctx)
}

func (r *TokenRefresher) isRefreshable() bool {
	return r.client.config.Authentication.isRefreshable()
}

// Starts a blocking process to monitor if the token stops working for reasons not due to errors
// (i.e exceeding the Max TTL) if so, it will re-authenticate with a new token & restart renewer processes
func (r *TokenRefresher) startRefresher(ctx context.Context) {
	for {
		select {
		case <-r.refreshCh:
			{
				r.renewer.Stop()
				if !r.isRefreshable() {
					break
				}

				err := r.client.Authenticate()
				if err != nil {
					logger.WithContext(ctx).Errorf("Could not get a new token: %v")
					break
				}
				logger.WithContext(ctx).Infof("%s token refreshed", renewerDescription)

				r.renewer, err = r.client.GetClientTokenRenewer()
				if err != nil {
					break
				}
				go r.renewer.Renew()
				go r.MonitorRenew(ctx)
			}
		}
	}
}

func (r *TokenRefresher) Stop() {
	r.renewer.Stop()
}

func (r *TokenRefresher) MonitorRenew(ctx context.Context) {
	for {
		select {
		case err := <-r.renewer.DoneCh():
			if err != nil {
				logger.WithContext(ctx).Errorf("%s renewer failed %v", renewerDescription, err)
			}

			r.refreshCh <- err
			logger.WithContext(ctx).Infof("%s renewer stopped", renewerDescription)
			break
		case renewal := <-r.renewer.RenewCh():
			logger.WithContext(ctx).Infof("%s successfully renewed at %v", renewerDescription, renewal.RenewedAt)
		}
	}
}
