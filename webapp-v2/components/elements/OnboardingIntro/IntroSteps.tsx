import Image from 'next/image';
import Warrior from 'public/img/man.svg';
import ProtectedServerSvg from 'public/img/protected_server.svg';

export const IntroStep0: React.FC = () => {
  return (
    <>
      <div className="space-y-4 md:pt-24 md:w-1/2">
        <h1 className="text-2xl lg:text-4xl font-bold uppercase">Welcome, security warrior</h1>
        <p className="text-xl lg:text-3xl md:w-3/4">
          We’re glad you are here.
          <br />
          Let’s get started.
        </p>
      </div>
      <div className="relative min-w-64 md:w-1/2 m-12 flex-1 base block">
        <Image src={Warrior} layout="fill" objectPosition="left" objectFit="contain" />
      </div>
    </>
  );
};

export const IntroStep1: React.FC = () => {
  return (
    <>
      <div className="relative min-w-64 md:w-1/2 m-12 flex-1 base block">
        <Image src={ProtectedServerSvg} layout="fill" objectPosition="left" objectFit="contain" />
      </div>
      <div className="space-y-4 self-center mb-20 md:w-1/2">
        <h1 className="text-2xl lg:text-4xl font-bold uppercase">Safety first</h1>
        <p className="text-xl lg:text-3xl md:w-3/4">
          immune Guard <span className="font-bold">protects your servers and edge devices</span> by
          detecting unauthorized changes from hackers and malware.
        </p>
      </div>
    </>
  );
};

export const IntroStep2: React.FC = () => {
  return (
    <>
      <div className="relative min-w-64 md:w-1/2 m-12 flex-1 base block">
        <Image src={Warrior} layout="fill" objectPosition="left" objectFit="contain" />
      </div>
      <div className="space-y-4 md:pt-24 md:w-1/2">
        <h1 className="text-2xl lg:text-4xl font-bold uppercase">Up to date</h1>
        <p className="text-xl lg:text-2xl md:w-3/4">
          The immune Guard Agent app registers your device with the immune Guard cloud service and{' '}
          <span className="font-bold">
            periodically transfers data about the device’s security.
          </span>
          <br />
          <br />
          You can see all your devices and whether they’re secure on the immune Guard dashboard. We
          can notify you via email or otherwise whenever the security status changes.
        </p>
      </div>
    </>
  );
};
