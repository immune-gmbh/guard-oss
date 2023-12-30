import classNames from 'classnames';
import HelperGate from 'components/containers/HelperGate/HelperGate';
import SidebarNavigation from 'components/elements/Navigations/SidebarNavigation';
import { useSession } from 'hooks/useSession';
import { SelectedDevicesProvider } from 'provider/SelectedDevicesProvider';

type IDashboardLayoutProps = React.HTMLProps<HTMLDivElement> & { mutateFn?: () => void };

export default function DashboardLayout({
  mutateFn,
  ...props
}: IDashboardLayoutProps): JSX.Element {
  const { session } = useSession();

  const { className, children, ...rest } = props;

  if (!session) {
    return null;
  }

  return (
    <SelectedDevicesProvider>
      <HelperGate mutateFn={mutateFn}>
        <div className="w-full flex items-stretch bg-cell-corner bg-no-repeat bg-right-bottom">
          <SidebarNavigation />
          <div className="w-content min-h-screen flex flex-col flex-1 overflow-auto">
            <div className={classNames('p-8 flex-1 max-w-7xl', className)} {...rest}>
              {/* implement something to show an info when there is no subscription or no free_credits */}
              {children}
            </div>
            <div id="actionbar-slot" className="sticky left-0 bottom-0 bg-purple-500"></div>
          </div>
        </div>
      </HelperGate>
    </SelectedDevicesProvider>
  );
}
