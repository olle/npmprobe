NPM Probe
=========

A simple script that does an exhaustive check for for compromised installed
`npm` packages - from an authored list.

## Quickstart

Simply probes for all packages and versions in the list of `compromised.txt`
packages.

```sh
> ./npmprobe compromised.txt ## do check
[OK]   @operato/help@operato/help@9.0.36 not present in any files
[OK]   @operato/help@operato/help@9.0.37 not present in any files
[OK]   @operato/help@operato/help@9.0.38 not present in any files
...
```

Probing may take a while and the results are output to the console, so that the
user can take further actions. The found compromised package and version is
displayed in the output. For example (not actual compromised package):

```sh
> npmprobe % ./npmprobe.sh compromised.txt
[FOUND] etag@1.8.1 in the following package.json files:
	/home/me/Development/project/docs/node_modules/vite/package.json:
	/home/me/Development/project/docs/node_modules/vitepress/node_modules/vite/package.json:
	/home/me/Development/project/docs/vitepress/node_modules/vite/package.json:
	/home/me/Development/project/node_modules/express/package.json:
	/home/me/Development/project/node_modules/send/package.json:
	/home/me/Development/project/node_modules/vite/package.json:
	/home/me/Development/lib/amqp-support/node_modules/vite/package.json:
	/home/me/Development/mr/mr-order-storage/node_modules/vite/package.json:
	/home/me/Development/md/backend/frontend/node_modules/vite/package.json:
	/home/me/Development/other/docs/node_modules/vite/package.json:
	/home/me/Development/other/node_modules/express/package.json:
	/home/me/Development/other/node_modules/send/package.json:
	/home/me/Development/other/src/rr/node_modules/vite/package.json:
	/home/me/Development/str/str/node_modules/vite/package.json:
	/home/me/Development/OpenSource/query-response-spring-amqp/node_modules/vite/package.json:
	/home/me/Development/OpenSource/query-response-spring-amqp/ui-frontend/node_modules/vite/package.json:
	/home/me/Development/OpenSource/spring-boot-vue-vite-mpa/node_modules/vite/package.json:
	/home/me/Development/OpenSource/spring-boot-vue-vite-mpa/src/main/vue/apps/one/node_modules/vite/package.json:
```

## Adding new packages

Simply edit the `compromised.txt` file and prepare it for publishing using the
provided `Makefile`.

```sh
> make ## fixes the package list file
```

Now commit the changes and either create a merge request or simply push it to
the central repository.

Happy hacking!