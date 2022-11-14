package openid

//TODO: success handler
// implement ordered, it should be before TokenRevokeSuccessHandler
// it should check post_logout_redirect_uri and id_token_hint and client_id
type OidcSuccessHandler struct {
}

//TODO: conditional handler
//it should only do something if post_logout_redirect_uri is there, in which case it checks the id_token_hint
