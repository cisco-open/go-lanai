# +short        client is authorized to access health details
# +desc         Example of checking if client is allowed to access health details
package actuator

import data.oauth2.has_scope

# +desc client with
allow_health_details {
    has_scope("admin")
}