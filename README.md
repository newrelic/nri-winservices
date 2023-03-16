<a href="https://opensource.newrelic.com/oss-category/#community-plus"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Community_Plus.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Plus.png"><img alt="New Relic Open Source community plus project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Community_Plus.png"></picture></a>

# New Relic Windows Services integration
![](https://github.com/newrelic/nri-winservices/workflows/PullRequestAndMergeMaster/badge.svg)
![](https://github.com/newrelic/nri-winservices/workflows/Release/badge.svg)

New Relic's Windows Services integration collects data from the services running on your Windows hosts into our platform. You can check the state and start mode of each service, find out which hosts are running a service, add services to workloads, set up alerts for services, and more.
 
For information on how to use and configure the Windows services integration, [read the official documentation](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/windows-services-integration). 
 
> Our integration is bundled with the [Windows agent](/docs/infrastructure/install-configure-manage-infrastructure/windows-installation/install-infrastructure-windows-server-using-msi-installer): if you are already monitoring Windows hosts on New Relic, you just need to enable the integration to get Windows services data into our platform.

# Architecture

To get data from Windows, the Windows services integration uses a reduced version of the [Prometheus exporter for 
Windows](https://github.com/prometheus-community/windows_exporter), which exposes Prometheus metrics on the port specified in the agent configuration. The integration collects these metrics, transforms them into entities, filters them, and then sent them to New Relic. 

![The Windows services integration collects Windows Management Instrumentation  (WMI) data using the Windows Prometheus exporter. It then transforms and filters the data before sending it to New Relic.](https://docs.newrelic.com/sites/default/files/thumbnails/image/WindowsServicesOHI.png)

## Installation

This integration comes bundled with New Relic's Windows infrastructure agent. It's not enabled by default. For installation and configuration instructions, [see the official documentation](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/windows-services-integration#install).

## Building

To build the integration, run `win_build.ps1`. This PowerShell script takes care of building the project and the supported version of the integration, placing the binaries in `/target/bin`.

> Note that only Windows is supported.

## Testing

Once built, the integration can be tested running `nri-winservices.exe`, which is in the `./target/bin` directory, using the config file in `./test/config.yml`. The command spins up automatically the exporter with the provided configuration. 

```powershell
PS .\nri-winservices.exe -config_path "../../../test/config.yml"
```

## Changelog

Changelog of releases is create by running `git-chglog  --next-tag v0.0.0`. 

Commit messages not following [the conventional commits pattern](https://www.conventionalcommits.org/en/v1.0.0/)  (es: `type(scope): what I have changed`) will be not included in the Changelog.

## Support

Should you need assistance with New Relic products, you are in good hands with several support diagnostic tools and support channels.

>This [troubleshooting framework](https://discuss.newrelic.com/t/troubleshooting-frameworks/108787) steps you through common troubleshooting questions.

>New Relic offers NRDiag, [a client-side diagnostic utility](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics) that automatically detects common problems with New Relic agents. If NRDiag detects a problem, it suggests troubleshooting steps. NRDiag can also automatically attach troubleshooting data to a New Relic Support ticket. Remove this section if it doesn't apply.

If the issue has been confirmed as a bug or is a feature request, file a GitHub issue.

**Support Channels**

* [New Relic Documentation](https://docs.newrelic.com): Comprehensive guidance for using our platform
* [New Relic Community](https://discuss.newrelic.com): The best place to engage in troubleshooting questions
* [New Relic Developer](https://developer.newrelic.com/): Resources for building a custom observability applications
* [New Relic University](https://learn.newrelic.com/): A range of online training for New Relic users of every level
* [New Relic Technical Support](https://support.newrelic.com/) 24/7/365 ticketed support. Read more about our [Technical Support Offerings](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/support-plan).

## Privacy

At New Relic we take your privacy and the security of your information seriously, and are committed to protecting your information. We must emphasize the importance of not sharing personal data in public forums, and ask all users to scrub logs and diagnostic information for sensitive information, whether personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or identifiable individual, including, for example, your name, phone number, post code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.

## License

nri-winservices is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
