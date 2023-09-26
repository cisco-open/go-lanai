package testservice_api

import data.mocks
import future.keywords

### Tests ###

# Test: GET /apim/api/v1/keys
test_get_keys_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/apim/api/v1/keys")}
}

test_get_keys_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/apim/api/v1/keys")}
}

# Test: DELETE /my/api/v1/controllerResponsesTest
test_delete_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("DELETE", "/my/api/v1/controllerResponsesTest")}
}

test_delete_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("DELETE", "/my/api/v1/controllerResponsesTest")}
}

# Test: GET /my/api/v1/controllerResponsesTest
test_get_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v1/controllerResponsesTest")}
}

test_get_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v1/controllerResponsesTest")}
}

# Test: HEAD /my/api/v1/controllerResponsesTest
test_head_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("HEAD", "/my/api/v1/controllerResponsesTest")}
}

test_head_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("HEAD", "/my/api/v1/controllerResponsesTest")}
}

# Test: PATCH /my/api/v1/controllerResponsesTest
test_patch_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PATCH", "/my/api/v1/controllerResponsesTest")}
}

test_patch_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PATCH", "/my/api/v1/controllerResponsesTest")}
}

# Test: PUT /my/api/v1/controllerResponsesTest
test_put_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PUT", "/my/api/v1/controllerResponsesTest")}
}

test_put_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PUT", "/my/api/v1/controllerResponsesTest")}
}

# Test: TRACE /my/api/v1/controllerResponsesTest
test_trace_controller_responses_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("TRACE", "/my/api/v1/controllerResponsesTest")}
}

test_trace_controller_responses_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("TRACE", "/my/api/v1/controllerResponsesTest")}
}

# Test: DELETE /my/api/v1/requestBodyTests/{id}
test_delete_request_body_tests_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("DELETE", "/my/api/v1/requestBodyTests/{id}")}
}

test_delete_request_body_tests_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("DELETE", "/my/api/v1/requestBodyTests/{id}")}
}

# Test: GET /my/api/v1/requestBodyTests/{id}
test_get_request_body_tests_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v1/requestBodyTests/{id}")}
}

test_get_request_body_tests_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v1/requestBodyTests/{id}")}
}

# Test: PATCH /my/api/v1/requestBodyTests/{id}
test_patch_request_body_tests_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PATCH", "/my/api/v1/requestBodyTests/{id}")}
}

test_patch_request_body_tests_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PATCH", "/my/api/v1/requestBodyTests/{id}")}
}

# Test: POST /my/api/v1/requestBodyTests/{id}
test_post_request_body_tests_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("POST", "/my/api/v1/requestBodyTests/{id}")}
}

test_post_request_body_tests_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("POST", "/my/api/v1/requestBodyTests/{id}")}
}

# Test: PUT /my/api/v1/requestBodyTests/{id}
test_put_request_body_tests_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PUT", "/my/api/v1/requestBodyTests/{id}")}
}

test_put_request_body_tests_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PUT", "/my/api/v1/requestBodyTests/{id}")}
}

# Test: DELETE /my/api/v1/testpath/{scope}
test_delete_testpath_scope_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("DELETE", "/my/api/v1/testpath/{scope}")}
}

test_delete_testpath_scope_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("DELETE", "/my/api/v1/testpath/{scope}")}
}

# Test: GET /my/api/v1/testpath/{scope}
test_get_testpath_scope_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v1/testpath/{scope}")}
}

test_get_testpath_scope_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v1/testpath/{scope}")}
}

# Test: PATCH /my/api/v1/testpath/{scope}
test_patch_testpath_scope_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PATCH", "/my/api/v1/testpath/{scope}")}
}

test_patch_testpath_scope_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PATCH", "/my/api/v1/testpath/{scope}")}
}

# Test: POST /my/api/v1/testpath/{scope}
test_post_testpath_scope_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("POST", "/my/api/v1/testpath/{scope}")}
}

test_post_testpath_scope_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("POST", "/my/api/v1/testpath/{scope}")}
}

# Test: GET /my/api/v1/uuidtest/{id}
test_get_uuidtest_id_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v1/uuidtest/{id}")}
}

test_get_uuidtest_id_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v1/uuidtest/{id}")}
}

# Test: GET /my/api/v2/testArrayUUID
test_get_test_array_uuid_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v2/testArrayUUID")}
}

test_get_test_array_uuid_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v2/testArrayUUID")}
}

# Test: PUT /my/api/v2/testArrayUUID
test_put_test_array_uuid_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("PUT", "/my/api/v2/testArrayUUID")}
}

test_put_test_array_uuid_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("PUT", "/my/api/v2/testArrayUUID")}
}

# Test: GET /my/api/v2/testpath
test_get_testpath_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v2/testpath")}
}

test_get_testpath_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v2/testpath")}
}

# Test: GET /my/api/v3/anotherTest
test_get_another_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("GET", "/my/api/v3/anotherTest")}
}

test_get_another_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("GET", "/my/api/v3/anotherTest")}
}

# Test: POST /my/api/v3/anotherTest
test_post_another_test_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("POST", "/my/api/v3/anotherTest")}
}

test_post_another_test_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("POST", "/my/api/v3/anotherTest")}
}

# Test: POST /my/api/v4/testAPIThatDoesntUseAnyImports
test_post_test_api_that_doesnt_use_any_imports_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("POST", "/my/api/v4/testAPIThatDoesntUseAnyImports")}
}

test_post_test_api_that_doesnt_use_any_imports_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("POST", "/my/api/v4/testAPIThatDoesntUseAnyImports")}
}

# Test: POST /my/api/v4/testRequestBodyUsingARefWithAllOf
test_post_test_request_body_using_a_ref_with_all_of_allowed if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.allowed_auth, "request": mocks_request("POST", "/my/api/v4/testRequestBodyUsingARefWithAllOf")}
}

test_post_test_request_body_using_a_ref_with_all_of_denied if {
	# TODO Implement test case and mocks
	allow_api with input as {"auth": mocks.denied_auth, "request": mocks_request("POST", "/my/api/v4/testRequestBodyUsingARefWithAllOf")}
}

### Mocks ###
mocks_request(method, path) = req if {
	req := {
		"endpoint_id": "env",
		"header": {
			"Authorization": ["Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImRldi0wIiwidHlwIjoiSldUIn0.eyJhdWQiOiJuZnYtYXBpIiwiY2xpZW50X2lkIjoibmZ2LXNlcnZpY2UiLCJlbWFpbCI6Im5vcmVwbHlAY2lzY28uY29tIiwiZXhwIjoxNjkxMTAwMjE4LCJmaXJzdE5hbWUiOiJTdXBlciIsImlhdCI6MTY5MTA5MTIxOCwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo4OTAwL2F1dGgiLCJqdGkiOiI1NGEzYjhlZC03NDBiLTRjOTctOGQ2OS00YWFkNjMzM2I1NmIiLCJsYXN0TmFtZSI6IlVzZXIiLCJuYmYiOjE2OTEwOTEyMTgsInJvbGVzIjpbIlNVUEVSVVNFUiJdLCJzY29wZSI6WyJyZWFkIiwid3JpdGUiLCJhZG1pbiJdLCJzdWIiOiJzdXBlcnVzZXIiLCJ0ZW5hbnRJZCI6IjdiMzkzNGZjLWVkYzQtNGExYy05MjQ5LTNkYzcwNTVlYjEyNCIsInVzZXJfbmFtZSI6InN1cGVydXNlciJ9.BkA_g1Eueu4l_dJYGB4_0_qZugQSslSG1DgldXUA3R1tEqoc6k5nHclBtRuTTEuwGRU8IUuxTnPaY_2di4MKXysw5PgqxwlcB7RozEvD51cB2jqyIMRW_FHfwm3dH4Im5-pK1lG5oTYFPyRHCAZ506S7mOlcqizvyT_IMiWwInyKtpRTyN0HFLbhkcQQodFfS-8OlFRGmOPctozr0f586Xaiw4nMwEIO5J7KzjLhIVK_hKEglAdjA0tsD118Nvn3P_KlVNYJrXdpdHpgVzjdWHKprUlGG6MQQ3EG3fjBGCu5pjbVmNCfd5zSAVUZ1cj-IudhyX4mH1pzz0daTHp1Aw"],
			"Content-Type": ["application/json"],
		},
		"method": method,
		"path": path,
	}
}
