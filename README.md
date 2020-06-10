[![New Relic Experimental header](https://github.com/newrelic/open-source-office/raw/master/examples/categories/images/Experimental.png)](https://github.com/newrelic/open-source-office/blob/master/examples/categories/index.md#category-new-relic-experimental)

# nri-winservices
![](https://github.com/newrelic/nri-winservices/workflows/PullRequestAndMergeMaster/badge.svg)
![](https://github.com/newrelic/nri-winservices/workflows/Release/badge.svg)

New Relic's Windows Services integration collects data from the services running on your Windows hosts into our platform. You can check the state, status, and start mode of each service, find out which hosts are running a service, add services to 
 workloads, set up alerts for services, and more.
 
 For information on how to use and configure the Windows services integration, [read the official documentation](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/windows-services-integration). 
 
> Our integration is bundled with the [Windows agent](/docs/infrastructure/install-configure-manage-infrastructure/windows-installation/install-infrastructure-windows-server-using-msi-installer): if you are already monitoring Windows hosts on New Relic, you just 
need to enable the integration to get Windows services data into our platform.

# Architecture

To get data from Windows, the Windows services integration uses a reduced version of the [Prometheus exporter for 
Windows](github.com/prometheus-community/windows_exporter), which exposes Prometheus metrics on the port specified in the agent configuration. The integration collects these metrics, transforms them into entities, filters them, and then sent them to New Relic. 

![The Windows services integration collects Windows Management Instrumentation  (WMI) data using the Windows Prometheus exporter. It then transforms and filters the data before sending it to New Relic.](https://docs.newrelic.com/sites/default/files/thumbnails/image/WindowsServicesOHI.png)

## Install

This integration comes bundled with New Relic's Windows infrastructure agent. It's not enabled by default. For installation and configuration instructions, [see the official documentation](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/windows-services-integration#install).

## Build

To build the integration, run `win_build.ps1`. This PowerShell script takes care of building the project and the supported version of the integration, placing the binaries in `/target/bin`.

> Note that only Windows is supported.

## Testing

Once built, the integration can be tested running `nri-winservices.exe`, which is in the `./target/bin` directory, using the config file in `./test/config.yml`. The command spins up automatically the exporter with the provided configuration. 

```powershell
PS .\nri-winservices.exe -config_path "../../../test/config.yml"
```

## Support

This project is not supported yet and it is currently under development.

## Contributing
Contributions to improve `nri-winservices` are encouraged! Keep in mind when you submit your pull request, you'll need to
 sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
To execute our corporate CLA, which is required if your contribution is on behalf of a company, or if you have any
 questions, please drop us an email at open-source@newrelic.com.

## License
nri-winservices is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
