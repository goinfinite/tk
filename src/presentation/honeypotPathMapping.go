package tkPresentation

type honeypotPayloadSpec struct {
	urlPath     string
	binFileName string
	mimeType    string
}

var honeypotPayloadSpecs = []honeypotPayloadSpec{
	{"/.env", "env.bin", "text/plain"},
	{
		"/wp-config.php",
		"wp-config.php.bin",
		"application/x-httpd-php",
	},
	{
		"/wp-config.php.bak",
		"wp-config.php.bak.bin",
		"application/x-httpd-php",
	},
	{
		"/config.php",
		"config.php.bin",
		"application/x-httpd-php",
	},
	{
		"/backup.sql",
		"backup.sql.bin",
		"application/sql",
	},
	{
		"/backup.zip",
		"backup.zip.bin",
		"application/zip",
	},
	{
		"/.git/config",
		"git-config.bin",
		"text/plain",
	},
	{
		"/.aws/credentials",
		"aws-credentials.bin",
		"text/plain",
	},
	{
		"/actuator/env",
		"actuator-env.json.bin",
		"application/json",
	},
	{
		"/actuator/configprops",
		"actuator-configprops.json.bin",
		"application/json",
	},
	{
		"/server-status",
		"server-status.html.bin",
		"text/html",
	},
	{
		"/phpmyadmin/index.php",
		"phpmyadmin-index.php.bin",
		"text/html",
	},
	{
		"/admin.php",
		"admin.php.bin",
		"text/html",
	},
	{
		"/administrator/index.php",
		"administrator-index.php.bin",
		"text/html",
	},
	{
		"/login.php",
		"login.php.bin",
		"text/html",
	},
	{
		"/shell.php",
		"shell.php.bin",
		"application/x-httpd-php",
	},
	{
		"/cmd.php",
		"cmd.php.bin",
		"application/x-httpd-php",
	},
	{
		"/test.php",
		"test.php.bin",
		"application/x-httpd-php",
	},
	{
		"/.htaccess",
		"htaccess.bin",
		"text/plain",
	},
	{
		"/web.config",
		"web.config.bin",
		"text/xml",
	},
	{
		"/robots.txt",
		"robots.txt.bin",
		"text/plain",
	},
	{
		"/sitemap.xml",
		"sitemap.xml.bin",
		"application/xml",
	},
	{
		"/debug.php",
		"debug.php.bin",
		"application/x-httpd-php",
	},
	{
		"/info.php",
		"info.php.bin",
		"application/x-httpd-php",
	},
	{
		"/console",
		"console.html.bin",
		"text/html",
	},
}

func findPayloadSpec(
	interceptPath string,
) *honeypotPayloadSpec {
	for specIdx := range honeypotPayloadSpecs {
		if honeypotPayloadSpecs[specIdx].urlPath ==
			interceptPath {
			return &honeypotPayloadSpecs[specIdx]
		}
	}
	return nil
}
