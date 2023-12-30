import TabNavigation from 'components/containers/TabNavigation/TabNavigation';
import CopyPasteBox from 'components/elements/CopyPasteBox/CopyPasteBox';
import Headline from 'components/elements/Headlines/Headline';
import NoDevices from 'components/elements/NoDevices/NoDevices';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import { useConfig } from 'hooks/config';
import { useDashboard } from 'hooks/dashboard';
import { useSession } from 'hooks/useSession';
import _ from 'lodash';
import getConfig from 'next/config';
import Head from 'next/head';
import { useMemo } from 'react';
import { handleLoadError } from 'utils/errorHandling';
import { SerializedMembership } from 'utils/types';

const { publicRuntimeConfig } = getConfig();

export function InstallDeviceComponent({
  currentMembership,
}: {
  currentMembership: SerializedMembership;
}): JSX.Element {
  const { config } = useConfig();

  const agents = useMemo(() => {
    if (config && config.agentUrls) {
      return Object.entries(config.agentUrls).map(([agent, url]) => {
        return { agent, url, filename: _.last(url.split('/')) };
      });
    }
    return [];
  }, [config]);

  const step2Explanation = `This will enroll your server by registering it’s unique hardware ID with the
  immune Guard cloud and establishing cryptographic keys inside your server’s
  trust anchor. After the enrollment was successful the device will attest
  itself by sending a signed statement of its configuration, identity and
  checksums of the boot chain. When the attestation is done the device will be
  listed on the device screen.`;
  const serverUrl = publicRuntimeConfig.isEdge
    ? ` --server ${publicRuntimeConfig.hosts.apiSrv}/v2`
    : '';

  const kernelModuleURL = `https://packages.immune.app/linux-kernel-dkms/flash_mmap-1.0.dkms.tar.gz`;

  return (
    <>
      <Head>
        <title>immune Guard | Device Registration</title>
      </Head>
      <div className="flex gap-y-4 flex-col">
        <Headline as="h1">Device Registration</Headline>
        <p>
          A minimal agent application is required to collect data from your device and share it with
          the immune Guard cloud servers.
        </p>
        <p>
          The agent is available for 64-bit Linux and Windows on x86. You must install the agent to
          add your device.
        </p>
        {agents.length > 0 && (
          <TabNavigation
            className="my-4"
            bodyClassNames="bg-gray-100 p-6"
            navigationPoints={[
              {
                title: 'Ubuntu / Debian',
                component: (
                  <div className="flex flex-col gap-y-8">
                    <ol className="list-decimal">
                      <li>
                        Download and install the immune Guard agent deb package.
                        <CopyPasteBox
                          text={`wget ${agents[0].url}
sudo dpkg -i ${agents[0].filename}`}
                        />
                      </li>
                      <li>
                        Download and install the immune Guard kernel module using DKMS. This
                        optional step is required for BIOS/UEFI analysis.
                        <CopyPasteBox
                          text={`wget ${kernelModuleURL}
sudo apt-get install dkms
sudo dkms ldtarball flash_mmap-1.0.dkms.tar.gz
sudo dkms autoinstall`}
                        />
                      </li>
                      <li>
                        Enroll your system. {step2Explanation}
                        <CopyPasteBox
                          text={`sudo guard enroll${serverUrl} ${currentMembership.enrollmentToken}`}
                        />
                      </li>
                    </ol>
                  </div>
                ),
              },
              {
                title: 'Fedora / RHEL',
                component: (
                  <div className="flex flex-col gap-y-8">
                    <ol className="list-decimal">
                      <li>
                        Download and install the immune Guard agent rpm package.
                        <CopyPasteBox
                          text={`wget ${agents[1].url}
sudo rpm -iU ${agents[1].filename}`}
                        />
                      </li>
                      <li>
                        Download and install the immune Guard kernel module using DKMS. This
                        optional step is required for BIOS/UEFI analysis.
                        <CopyPasteBox
                          text={`wget ${kernelModuleURL}
sudo dnf install dkms
sudo dkms ldtarball flash_mmap-1.0.dkms.tar.gz
sudo dkms autoinstall`}
                        />
                      </li>
                      <li>
                        When Secure Boot is used run the following command, reboot your system and
                        enroll the MOK via the user interface.
                        <CopyPasteBox text="sudo mokutil --import /var/lib/dkms/mok.pub" />
                      </li>
                      <li>
                        Enroll your system. {step2Explanation}
                        <CopyPasteBox
                          text={`sudo guard enroll${serverUrl} ${currentMembership.enrollmentToken}`}
                        />
                      </li>
                    </ol>
                  </div>
                ),
              },
              {
                title: 'Generic GNU/Linux',
                component: (
                  <div className="flex flex-col gap-y-8">
                    <ol className="list-decimal">
                      <li>
                        Download and install the immune Guard agent standalone binary.
                        <CopyPasteBox
                          text={`sudo wget ${agents[2].url} -O /usr/bin/guard
sudo chmod +x /usr/bin/guard`}
                        />
                      </li>
                      <li>
                        Download and install the immune Guard kernel module using DKMS. This
                        optional step is required for BIOS/UEFI analysis.
                        <CopyPasteBox
                          text={`wget ${kernelModuleURL}
sudo dkms ldtarball flash_mmap-1.0.dkms.tar.gz
sudo dkms autoinstall`}
                        />
                      </li>
                      <li>
                        Enroll your system. {step2Explanation}
                        <CopyPasteBox
                          text={`sudo guard enroll${serverUrl} ${currentMembership.enrollmentToken}`}
                        />
                      </li>
                    </ol>
                  </div>
                ),
              },
              {
                title: 'Microsoft Windows',
                component: (
                  <div className="flex flex-col gap-y-8">
                    <ol className="list-decimal">
                      <li>
                        Download and install the immune Guard agent.
                        <CopyPasteBox text={agents[3].url} />
                      </li>
                      <li>
                        Enter the following enrollment token into the installer when asked.{' '}
                        {step2Explanation}
                        <CopyPasteBox text={`${currentMembership.enrollmentToken}`} />
                      </li>
                    </ol>
                  </div>
                ),
              },
            ]}
          />
        )}
        {agents?.length == 0 && <Spinner />}
      </div>
    </>
  );
}

export default function DevicesAdd(): JSX.Element {
  const { data, isLoading, isError } = useDashboard(true);

  const {
    session: { currentMembership },
  } = useSession();

  if (isError) return handleLoadError('Devices');
  if (isLoading) return <Spinner />;

  const countDevices = Object.values(data.deviceStats).reduce((prev, curr) => prev + curr, 0);

  return (
    <div className="space-y-16">
      {countDevices == 0 && <NoDevices />}
      <InstallDeviceComponent currentMembership={currentMembership} />
    </div>
  );
}

DevicesAdd.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
