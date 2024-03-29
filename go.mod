module github.com/sonatype-nexus-community/hashbrowns

go 1.14

require (
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/jarcoal/httpmock v1.0.5
	github.com/magiconair/properties v1.8.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/sonatype-nexus-community/go-sona-types v0.0.11
	github.com/sonatype-nexus-community/nancy v0.2.3
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	gopkg.in/ini.v1 v1.60.2 // indirect
)

// fix vulnerability: CVE-2021-38561 in golang.org/x/text@v0.3.3
// fix vulnerability: CVE-2022-32149 in golang.org/x/text v0.3.7
replace golang.org/x/text => golang.org/x/text v0.3.8
