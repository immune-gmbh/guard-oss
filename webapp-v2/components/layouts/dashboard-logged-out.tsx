import Image from 'next/image';
import Logo from 'public/img/logo-white.svg';

const DashboardLoggedOut: React.FC = ({ children }) => (
  <div className="bg-purple-500 h-full min-h-screen items-center text-white bg-cell-dashboard-bg bg-cell-dashboard bg-left-bottom bg-no-repeat">
    <div className="w-full px-12 flex flex-col pt-12 my-0 h-full xl:w-content xl:mx-auto xl:px-0 min-h-screen">
      <Image src={Logo.src || '/img/logo-white.svg'} width={Logo.width || 220} height={70} />
      <div className="flex flex-1 flex-col">{children}</div>
    </div>
  </div>
);
export default DashboardLoggedOut;
