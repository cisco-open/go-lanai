<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta http-equiv='cache-control' content='no-cache, no-store, max-age=0, must-revalidate'>
		<meta http-equiv='expires' content='0'>
		<meta http-equiv='pragma' content='no-cache'>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title></title>
		<link rel="stylesheet" href="${rc.contextPath}/webjars/css/vendor.css?1619857593316">
		<link rel="icon" type="image/png" href="${rc.contextPath}/webjars/images/favicon.png">
		<link rel="stylesheet" id="msx-root-css" href="${rc.contextPath}/webjars/css/boot.css?1619857593316">
		<link rel="stylesheet" id="msx-extension-css" href="/css/app-custom.css?1619857593316">
	</head>
	<body ng-app="skPlatformBoot" style="overflow-y: auto;" class="layout-fixed vms-body sk-font-standard-text">
		<div ui-view="content"></div>
		<!-- Main View -->
		<script>
			// Create a globalised dynamic csrf token for this login flow.
			var csrfToken = "${(_csrf.token)!""}";
			var csrfParameterName = "${(_csrf.parameterName)!""}";
			var contextPath = "${rc.contextPath}"
			<#assign error = SPRING_SECURITY_LAST_EXCEPTION!SPRING_SECURITY_403_EXCEPTION!({})/>
			<#if RequestParameters.error?? && error.message?has_content>
				var authErrors = "${error.message!}";
			<#else>
				var authErrors = "";
			</#if>
			var authEndpoint = "${PasswordLoginEndpoints.formLoginProcess}";
			var optVerifyEndpoint = "${PasswordLoginEndpoints.otpVerifyProcess!}";
			var optResendEndpoint = "${PasswordLoginEndpoints.otpVerifyResend!}";
			var MSX_VERSION = <#if MSXVersion??>"${MSXVersion}"<#else>"4.0.0"</#if>;
			var MSX_LOGIN_VERSION="4.0.0";
			
			var dojoConfig = {
				baseUrl: "${rc.contextPath}/webjars/vendor",
				async: true,
				extraLocale: [
				]
			};
		</script>
		<script src="${rc.contextPath}/webjars/js/dojo-min.js?1619857593316"></script>
		<script src="${rc.contextPath}/webjars/js/vendor.js?1619857593316"></script>
		<script src="${rc.contextPath}/webjars/js/boot.js?1619857593316"></script>
	</body>
</html>
