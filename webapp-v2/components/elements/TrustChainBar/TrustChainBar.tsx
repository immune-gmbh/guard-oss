import classNames from 'classnames';
import useTranslation from 'next-translate/useTranslation';
import { Fragment } from 'react';
import { ApiSrv } from 'types/apiSrv';

import TrustChainIcon, { VERDICT_STATUS_STEPS } from './TrustChainIcon';
import TrustChainLine from './TrustChainLine';

interface ITrustChainBarWithVerdict {
  verdict: ApiSrv.Verdict;
  unknown?: boolean;
}

interface ITrustChainBarWithUnknown {
  verdict?: ApiSrv.Verdict;
  unknown: boolean;
}

export default function TrustChainBar({
  verdict,
  unknown,
}: ITrustChainBarWithVerdict | ITrustChainBarWithUnknown): JSX.Element {
  const { t } = useTranslation('common');
  let trustedUntilThisStep = unknown ? false : true;

  if (!verdict) {
    return null;
  }

  return (
    <div className={unknown ? 'opacity-25' : null}>
      <section className="h-auto space-y-4">
        <div className="grid grid-cols-boot-status-bar grid-rows-2 gap-y-4 items-center">
          {Object.keys(VERDICT_STATUS_STEPS).map(
            (step: keyof typeof VERDICT_STATUS_STEPS, index) => {
              const unsupported = verdict[step] === 'unsupported';
              const trusted = unknown ? true : verdict[step] === 'trusted' || unsupported;
              const lastRound = index < Object.keys(VERDICT_STATUS_STEPS).length - 1;

              const elements = (
                <Fragment key={step}>
                  <TrustChainIcon
                    trusted={trusted}
                    unsupported={unsupported}
                    trustedUntilThisStep={trustedUntilThisStep}
                    icon={step}
                  />
                  <span
                    className={classNames('row-start-2 text-center self-start', {
                      'text-green-notification': trustedUntilThisStep && !unsupported,
                      'text-purple-500': !trustedUntilThisStep || unsupported,
                      'text-red-cta': trustedUntilThisStep && !trusted && !unsupported,
                    })}>
                    {t(step)}
                  </span>
                  <span className="row-start-2" />
                  {lastRound && (
                    <TrustChainLine
                      status={trustedUntilThisStep && trusted}
                      unsupported={unsupported}
                    />
                  )}
                </Fragment>
              );
              trustedUntilThisStep = trustedUntilThisStep && trusted;

              return elements;
            },
          )}
        </div>
      </section>
    </div>
  );
}
