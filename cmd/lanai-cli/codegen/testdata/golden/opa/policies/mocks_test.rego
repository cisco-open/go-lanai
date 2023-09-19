package mocks

### TODO Change mocks based on test cases

allowed_auth := {
	"user_id": "5b4fd9b0-5913-499e-a06a-e67f6feccaab",
	"permissions": [
		"MOCKED_PERMISSION1",
		"MOCKED_PERMISSION2",
		"MOCKED_PERMISSION3",
	],
	"username": "superuser",
	"tenant_id": "7b3934fc-edc4-4a1c-9249-3dc7055eb124",
	"provider_id": "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266",
	"roles": ["MOCKED_ROLE"],
	"accessible_tenants": ["7b3934fc-edc4-4a1c-9249-3dc7055eb124"],
	"client": {
		"client_id": "nfv-client",
		"scopes": [
			"read",
			"write",
			"admin",
		],
	},
}

denied_auth := {
	"user_id": "5b4fd9b0-5913-499e-a06a-e67f6feccaab",
	"permissions": ["MOCKED_PERMISSION"],
	"username": "regular",
	"tenant_id": "7b3934fc-edc4-4a1c-9249-3dc7055eb124",
	"provider_id": "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266",
	"roles": ["MOCKED_ROLE"],
	"accessible_tenants": ["7b3934fc-edc4-4a1c-9249-3dc7055eb124"],
	"client": {
		"client_id": "nfv-client",
		"scopes": [
			"read",
			"write",
		],
	},
}
