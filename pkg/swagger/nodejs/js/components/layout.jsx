/**
 * Copyright 2023 Cisco Systems, Inc. and its affiliates
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react"
import PropTypes from "prop-types"

export default class SsoStandaloneLayout extends React.Component {

    static propTypes = {
        ssoSelectors: PropTypes.object.isRequired,
        ssoActions: PropTypes.object.isRequired,
        errSelectors: PropTypes.object.isRequired,
        errActions: PropTypes.object.isRequired,
        specActions: PropTypes.object.isRequired,
        specSelectors: PropTypes.object.isRequired,
        layoutSelectors: PropTypes.object.isRequired,
        layoutActions: PropTypes.object.isRequired,
        getComponent: PropTypes.func.isRequired,
    }

    render() {
        let { getComponent, ssoSelectors, specSelectors, errSelectors } = this.props

        const Container = getComponent("Container")
        const Row = getComponent("Row")
        const Col = getComponent("Col")
        const Errors = getComponent("errors", true)

        const SsoTopBar = getComponent("SsoTopBar", true)
        const BaseLayout = getComponent("BaseLayout", true)
        const OnlineValidatorBadge = getComponent("onlineValidatorBadge", true)

        const loadingStatus = specSelectors.loadingStatus()
        const lastErr = errSelectors.lastError()
        const lastErrMsg = lastErr ? lastErr.get("message") : ""
        const authorizing = ssoSelectors.isAuthorizing();
        const authorized = ssoSelectors.isAuthorized();

        return (

            <Container className='swagger-ui' >
                { SsoTopBar ? <SsoTopBar /> : null }
                { !loadingStatus && authorizing &&
                <div className="info">
                    <div className="loading-container">
                        <div className="markdown">Authorization in Progress</div>
                    </div>
                    <div className="loading-container">
                        <div className="loading"></div>
                    </div>
                </div>
                }
                { !loadingStatus && !authorizing && !authorized &&
                <div className="info">
                    <div className="loading-container">
                        <Errors/>
                        <div className="markdown">Login is required to load API definition.</div>
                    </div>
                </div>
                }
                { loadingStatus === "loading" &&
                <div className="info">
                    <div className="loading-container">
                        <div className="loading"></div>
                    </div>
                </div>
                }
                { loadingStatus === "failed" &&
                <div className="info">
                    <div className="loading-container">
                        <h4 className="title">Failed to load API definition.</h4>
                        <Errors/>
                    </div>
                </div>
                }
                { loadingStatus === "failedConfig" &&
                <div className="info" style={{ maxWidth: "880px", marginLeft: "auto", marginRight: "auto", textAlign: "center" }}>
                    <div className="loading-container">
                        <h4 className="title">Failed to load remote configuration.</h4>
                        <p>{lastErrMsg}</p>
                    </div>
                </div>
                }
                { loadingStatus === "success" && <BaseLayout /> }
                <Row>
                    <Col>
                        <OnlineValidatorBadge />
                    </Col>
                </Row>
            </Container>
        )
    }

}