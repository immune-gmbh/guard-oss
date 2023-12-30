import { DotsVerticalIcon } from '@heroicons/react/solid';

export default function MenuListIcon(): JSX.Element {
  return (
    <>
      <DotsVerticalIcon height="15" />
      <DotsVerticalIcon height="15" className="-ml-dots-icon" />
      <DotsVerticalIcon height="15" className="-ml-dots-icon pr-4" />
    </>
  );
}
