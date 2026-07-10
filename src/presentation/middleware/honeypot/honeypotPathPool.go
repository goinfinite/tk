package tkPresentationMiddlewareHoneypot

import (
	"log/slog"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotPathPool struct {
	pathsByClass map[tkValueObject.HoneypotPathClass][]tkValueObject.UrlPath
}

func NewHoneypotPathPool() HoneypotPathPool {
	pathsByClass := map[tkValueObject.HoneypotPathClass][]tkValueObject.UrlPath{
		tkValueObject.HoneypotPathClassStaticVulnerability: buildPaths(
			staticVulnerabilityRawPaths,
		),
		tkValueObject.HoneypotPathClassBandwidthExhaust: buildPaths(
			bandwidthExhaustRawPaths,
		),
		tkValueObject.HoneypotPathClassAiTrap: buildPaths(
			aiTrapRawPaths,
		),
	}

	return HoneypotPathPool{pathsByClass: pathsByClass}
}

func (pool HoneypotPathPool) Size() int {
	totalCount := 0
	for _, paths := range pool.pathsByClass {
		totalCount += len(paths)
	}
	return totalCount
}

func (pool HoneypotPathPool) PathsByClass(
	class tkValueObject.HoneypotPathClass,
) []tkValueObject.UrlPath {
	return pool.pathsByClass[class]
}

func buildPaths(
	rawPaths []string,
) []tkValueObject.UrlPath {
	paths := make([]tkValueObject.UrlPath, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		path, err := tkValueObject.NewUrlPath(rawPath)
		if err != nil {
			slog.Debug(
				"HoneypotPathPoolPathValidationError",
				"rawPath", rawPath,
				"error", err,
			)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

var staticVulnerabilityRawPaths = []string{
	"/wp-config.php",
	"/.env",
	"/.env.local",
	"/.env.production",
	"/backup.sql",
	"/dump.sql",
	"/database.sql",
	"/db_backup.sql",
	"/admin/config.php",
	"/administrator/config.php",
	"/wp-admin/install.php",
	"/wp-login.php",
	"/xmlrpc.php",
	"/.git/config",
	"/.git/HEAD",
	"/.svn/entries",
	"/.htaccess",
	"/.htpasswd",
	"/server-status",
	"/server-info",
	"/phpinfo.php",
	"/info.php",
	"/test.php",
	"/debug.php",
	"/config.php",
	"/config.yml",
	"/config.yaml",
	"/config.json",
	"/config.ini",
	"/settings.php",
	"/settings.yml",
	"/application.yml",
	"/application.properties",
	"/web.config",
	"/crossdomain.xml",
	"/clientaccesspolicy.xml",
	"/.well-known/security.txt",
	"/robots.txt",
	"/sitemap.xml",
	"/.DS_Store",
	"/Thumbs.db",
	"/.bash_history",
	"/.ssh/id_rsa",
	"/etc/passwd",
	"/etc/shadow",
	"/proc/self/environ",
	"/proc/version",
	"/.dockerenv",
	"/Dockerfile",
	"/docker-compose.yml",
	"/.travis.yml",
	"/.circleci/config.yml",
	"/Jenkinsfile",
	"/Vagrantfile",
	"/composer.json",
	"/composer.lock",
	"/package.json",
	"/package-lock.json",
	"/yarn.lock",
	"/Gemfile",
	"/Gemfile.lock",
	"/requirements.txt",
	"/Pipfile",
	"/Pipfile.lock",
	"/go.sum",
	"/Cargo.toml",
	"/Cargo.lock",
	"/.npmrc",
	"/.yarnrc",
	"/.babelrc",
	"/tsconfig.json",
	"/webpack.config.js",
	"/vite.config.js",
	"/next.config.js",
	"/nuxt.config.js",
	"/.eslintrc.json",
	"/.prettierrc",
	"/.editorconfig",
	"/Makefile",
	"/CMakeLists.txt",
	"/build.gradle",
	"/pom.xml",
	"/.classpath",
	"/.project",
	"/.settings/org.eclipse.jdt.core.prefs",
}

var bandwidthExhaustRawPaths = []string{
	"/api/v1/stream/logs",
	"/api/v1/events",
	"/api/v1/metrics/export",
	"/api/v1/audit/trail",
	"/api/v1/telemetry/batch",
	"/api/v2/stream/events",
	"/api/v2/analytics/export",
	"/api/v2/monitoring/realtime",
	"/api/v2/reports/generate",
	"/api/v2/data/sync",
	"/api/internal/diagnostics",
	"/api/internal/healthcheck/verbose",
	"/api/internal/debug/trace",
	"/api/internal/profiling/cpu",
	"/api/internal/profiling/memory",
	"/api/v1/backups/download",
	"/api/v1/logs/aggregate",
	"/api/v1/logs/search",
	"/api/v1/logs/stream",
	"/api/v1/logs/export",
}

var aiTrapRawPaths = []string{
	"/api/v1/ai/models",
	"/api/v1/ai/completions",
	"/api/v1/ai/embeddings",
	"/api/v1/ai/fine-tunes",
	"/api/v1/ai/datasets",
	"/api/v2/ai/models",
	"/api/v2/ai/inference",
	"/api/v2/ai/training/jobs",
	"/api/v2/ai/evaluation",
	"/api/internal/ai/experiments",
	"/api/internal/ai/checkpoints",
	"/api/internal/ai/tokenizer",
	"/v1/chat/completions",
	"/v1/completions",
	"/v1/embeddings",
	"/v1/models",
	"/v1/images/generations",
	"/v1/audio/transcriptions",
	"/v1/audio/translations",
	"/v1/moderations",
}
