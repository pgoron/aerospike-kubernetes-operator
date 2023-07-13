package v1beta1

import "testing"

func TestLdapQueryEscaping(t *testing.T) {
	ldapConf := map[string]string{
		"disable-tls":              "false",
		"login-threads":            "64",
		"query-base-dn":            "DC=uad,DC=preprod,DC=example,DC=org",
		"query-user-dn":            "CN=serviceaccount,OU=ou,DC=dc,DC=preprod,DC=example,DC=org",
		"query-user-password-file": "/etc/aerospike/secret/ldap_password.txt",
		"role-query-pattern":       "(&(objectClass=group)(member=${dn}))",
		"server":                   "ldap://uad.preprod.example.org",
		"tls-ca-file":              "/etc/aerospike/secret/ldap_ca.crt",
		"user-query-pattern":       "(cn=${un})",
	}
	securityConf := map[string]interface{}{
		"ldap": ldapConf,
	}
	config := AerospikeConfigSpec{}
	config.Value = map[string]interface{}{
		"security": securityConf,
	}

	escapeLDAPConfiguration(config)
	if _, ok := config.Value["security"]; ok {
		securityConf := config.Value["security"].(map[string]interface{})
		if _, ok := securityConf["ldap"]; ok {
			ldapConf := securityConf["ldap"].(map[string]string)

			roleQueryPattern := ldapConf["role-query-pattern"]
			expectedRoleQueryPattern := "(&(objectClass=group)(member=\\${dn}))"
			if roleQueryPattern != expectedRoleQueryPattern {
				t.Errorf("role-query-pattern unproperly escaped, get %s, expected %s", roleQueryPattern, expectedRoleQueryPattern)
			}

			userQueryPattern := ldapConf["user-query-pattern"]
			expectedUserQueryPattern := "(cn=\\${un})"
			if userQueryPattern != expectedUserQueryPattern {
				t.Errorf("role-query-pattern unproperly escaped, get %s, expected %s", userQueryPattern, expectedUserQueryPattern)
			}
		} else {
			t.Error("ldap section lost")
		}
	} else {
		t.Error("security section lost")
	}
}
