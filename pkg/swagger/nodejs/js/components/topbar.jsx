import React, { cloneElement } from "react"
import PropTypes from "prop-types"
import Logo from "../../images/msx-dark-1.svg"

export default class SsoTopBar extends React.Component {

    static propTypes = {
        layoutActions: PropTypes.object.isRequired,
        ssoActions: PropTypes.object.isRequired,
        ssoSelectors: PropTypes.object.isRequired,
        specSelectors: PropTypes.object.isRequired,
        specActions: PropTypes.object.isRequired,
        getComponent: PropTypes.func.isRequired,
        getConfigs: PropTypes.func.isRequired,
        parameterName: PropTypes.object,
        parameterValue: PropTypes.object
    }

    constructor(props, context) {
        super(props, context)
        this.state = {
            url: props.specSelectors.url(),
            selectedIndex: 0,
            loadSpecAttempted: false
        }
    }

    componentWillReceiveProps(nextProps) {
        this.setState({
            url: nextProps.specSelectors.url()
        })
    }

    componentDidUpdate() {
        let { ssoSelectors, getConfigs } = this.props;

        if (!this.state.loadSpecAttempted && ssoSelectors.isAuthorized()) {
            const configs = getConfigs()
            const urls = configs.urls || []
            this.loadSpec(urls[this.state.selectedIndex].url)
        }
    }

    loadSpec = (url) => {
        let { ssoSelectors, specActions } = this.props;
        if (ssoSelectors.isAuthorized() ) {
            this.setState({loadSpecAttempted: true});
            specActions.updateUrl(url);
            specActions.download(url);
        } else {
            this.ssoAuthorize();
        }
    }

    ssoAuthorize = () => {
        if (this.props.ssoSelectors.ssoConfigs()) {
            this.props.ssoActions.ssoAuthorize(this.props);
        }
    }

    ssoRefreshToken = () => {
        if (this.props.ssoSelectors.ssoConfigs()) {
            this.props.ssoActions.accessTokenExpired();
        }
    }

    onUrlSelect = (e)=> {
        const url = e.target.value || e.target.href;
        this.setState({loadSpecAttempted: false});
        this.loadSpec(url)
        e.preventDefault()
    }

    onLoginClick = (e) => {
        this.ssoAuthorize();
        e.preventDefault()
    }

    onAuthorizeWithParams = (e) => {
        const form = e.target;
        const formData = new FormData(form);
        this.props.parameterName = formData.get("parameterName")
        this.props.parameterValue = formData.get("parameterValue")
        this.ssoAuthorize();
        e.preventDefault()
    }

    onLogoutClick = (e) => {
        e.preventDefault()
    }

    onRefreshClick = (e) => {
        this.ssoRefreshToken();
        e.preventDefault()
    }

    componentDidMount() {
        const configs = this.props.getConfigs()
        const urls = configs.urls || []

        if(urls && urls.length) {
            let targetIndex = this.state.selectedIndex
            const primaryName = configs["urls.primaryName"]
            if(primaryName)
            {
                urls.forEach((spec, i) => {
                    if(spec.name === primaryName)
                    {
                        this.setState({selectedIndex: i})
                        targetIndex = i
                    }
                })
            }

            this.loadSpec(urls[targetIndex].url);
        }
    }

    onFilterChange =(e) => {
        const {target: {value}} = e
        this.props.layoutActions.updateFilter(value)
    }

    render() {
        let { getComponent, ssoSelectors, specSelectors, getConfigs } = this.props
        const Button = getComponent("Button")
        const Link = getComponent("Link")

        const isLoading = specSelectors.loadingStatus() === "loading";
        const isAuthorized = ssoSelectors.isAuthorized();
        const hasRefreshToken = ssoSelectors.hasRefreshToken();
        const username = ssoSelectors.getFromTokenResponse("username");
        const tenantId = ssoSelectors.getFromTokenResponse("tenantId");

        var params = []
        if (ssoSelectors.ssoConfigs()) {
            params = ssoSelectors.ssoConfigs()["additionalParameters"]
        }

        let { url, urls } = getConfigs()

        const control = []

        if(!urls || !(urls instanceof Array)) {
            urls = [];
        }
        if (url) {
            urls.unshift({url, name: url})
        }

        const rows = []
        urls.forEach((link, i) => {
            rows.push(<option key={i} value={link.url}>{link.name}</option>)
        })

        control.push(
            <label className="select-label" htmlFor="select" style={{width: '20em', 'margin-right': '0.5em'}}>
                <select id="select" disabled={isLoading} onChange={ this.onUrlSelect } value={urls[this.state.selectedIndex].url}>
                    {rows}
                </select>
            </label>
        )

        return (
            <div className="topbar">
                <div className="wrapper">
                    <div className="topbar-wrapper">
                        <Link>
                            <Logo height={40} width={40} viewBox="0 0 1024 1024"/>
                        </Link>
                        { isAuthorized && hasRefreshToken && params &&
                            <div style={{color: '#fff'}}>{username}, Tenant {(tenantId) ? tenantId : "not selected" }</div>
                        }
                        <form className="download-url-wrapper" style={{visibility: 'hidden'}}>
                            {control.map((el, i) => cloneElement(el, { key: i }))}
                        </form>

                        { isAuthorized && hasRefreshToken &&
                            <>
                                <Button className="btn authorize" onClick={ this.onRefreshClick }>Refresh</Button>
                                { params &&
                                    <form onSubmit={this.onAuthorizeWithParams}>
                                        <label>
                                            &nbsp;&nbsp;
                                            <select name="parameterName">
                                                {params.map(function(item) {
                                                    return <option label={item.displayName} value={item.name}>{item.name}</option>;
                                                })}
                                            </select>
                                        </label>
                                        <label><input name="parameterValue"/></label>
                                        <button className="btn authorize" type="submit">Authorize</button>
                                    </form>
                                }
                            </>
                        }
                        { !isAuthorized &&
                            <Button className="btn authorize" onClick={ this.onLoginClick }>Login</Button>
                        }
                    </div>
                </div>
            </div>
        )
    }
}
