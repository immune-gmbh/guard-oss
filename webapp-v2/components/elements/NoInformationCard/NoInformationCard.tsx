import { InformationCircleIcon } from '@heroicons/react/outline';

export default function NoInformationCard(): JSX.Element {
  return (
    <section className="shadow-sm p-8 bg-red-critical text-purple-500">
      <div className="grid grid-cols-[30px,1fr] grid-rows-2 gap-x-4">
        <div className="col-start-1">
          <InformationCircleIcon className="w-8 h-8" />
        </div>
        <h3 className="col-start-2 font-bold text-2xl text-red-cta">No information yet</h3>
        <span className="col-start-2 row-start-2">
          We did not hear anything from this server yet. Please check back later.
        </span>
      </div>
    </section>
  );
}
