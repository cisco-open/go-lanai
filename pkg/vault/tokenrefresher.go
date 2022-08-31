package vault

import (
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
	"sync"
	"time"
)

// TokenRefresher performs renewal & refreshment of a client's token
// renewal can occur when a token's ttl is completed,
// refresh occurs when a token cannot be renewed (e.g max TTL is reached)
type TokenRefresher struct {
	client     Client
	renewer    *api.Renewer
	cancelFunc context.CancelFunc
	cancelLock sync.Mutex
}

const renewerDescription = "vault apiClient token"

func NewTokenRefresher(client *Client) *TokenRefresher {
	return &TokenRefresher{
		client: *client,
	}
}

// Start will begin the processes of token renewal & refreshing
func (r *TokenRefresher) Start(ctx context.Context) {
	r.cancelLock.Lock()
	defer r.cancelLock.Unlock()
	if r.cancelFunc != nil {
		return
	}
	ctx, r.cancelFunc = context.WithCancel(ctx)
	//this starts a background process to log the renewal events.
	go r.monitorRenew(ctx)
}

// Stop will stop the token renewal/refreshing processes
func (r *TokenRefresher) Stop() {
	r.cancelLock.Lock()
	defer r.cancelLock.Unlock()

	if r.cancelFunc != nil {
		r.cancelFunc()
		r.cancelFunc = nil
	}
}

func (r *TokenRefresher) isRefreshable() bool {
	return r.client.config.Authentication.isRefreshable()
}

// Starts a blocking process to monitor if the token stops being renewed
// If so, it will refresh the token (if refreshable) and restart renewing process
func (r *TokenRefresher) monitorRenew(ctx context.Context) {
	for {
		if r.renewer == nil {
			// If the token expires or if the lease is revoked
			// Sleep for some time and see if the token valid now (i.e if the token is recreated by vault)
			for {
				var err error
				if r.renewer, err = r.client.GetClientTokenRenewer(); err == nil {
					break
				}
				logger.WithContext(ctx).Debugf("%s unable to create token renewer, %v", renewerDescription, err)
				time.Sleep(5 * time.Minute)
			}
			// Starts a blocking process to periodically renew the token.
			go r.renewer.Start()
		}
		select {
		case renewal := <-r.renewer.RenewCh():
			logger.WithContext(ctx).Debugf("%s successfully renewed at %v", renewerDescription, renewal.RenewedAt)
		case err := <-r.renewer.DoneCh():
			r.renewer = nil
			switch {
			case err != nil && !r.isRefreshable():
				// When authentication is token, and if the token expires, we can't really do anything on the client side
				// Do not quit the renewer in the hopes that the token is recreated & we can resume
				logger.WithContext(ctx).Warnf("%s renewer stopped for non-refreshable authentication: %v", renewerDescription, err)
				break
			case err != nil:
				logger.WithContext(ctx).Infof("%s renewer stopped with error, will re-authenticate & restart: %v", renewerDescription, err)
			default:
				logger.WithContext(ctx).Debugf("%s renewer stopped, will re-authenticate & restart", renewerDescription)
			}

			err = r.client.Authenticate()
			if err != nil {
				logger.WithContext(ctx).Errorf("Could not get a new token: %v")
				break
			}

		case <-ctx.Done():
			r.renewer.Stop()
			r.renewer = nil
			return
		}
	}
}
