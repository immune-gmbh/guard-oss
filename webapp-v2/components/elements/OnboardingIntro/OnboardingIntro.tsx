import { Transition } from '@headlessui/react';
import Button from 'components/elements/Button/Button';
import ImmuneLink from 'components/elements/Button/Link';
import NextJsRoutes from 'generated/NextJsRoutes';
import { default as React, useState } from 'react';

import IntroStepIndicator from './IntroStepIndicator';
import { IntroStep0, IntroStep1, IntroStep2 } from './IntroSteps';

interface IOnboardingIntroProps {}

const ContentTransition = ({ show, children }): JSX.Element => (
  <Transition
    show={show}
    className="flex flex-1 flex-col md:flex-row p-8 md:p-4"
    enter="transition-all ease-in-out duration-1000 transform translate-x-0"
    enterFrom="opacity-0"
    enterTo="opacity-100"
    leave="transition-opacity duration-150"
    leaveFrom="opacity-100"
    leaveTo="opacity-0">
    {children}
  </Transition>
);

const OnboardingIntro: React.FC<IOnboardingIntroProps> = () => {
  const [currentStep, setCurrentStep] = useState(0);

  return (
    <>
      <ContentTransition show={currentStep == 0}>
        <IntroStep0 />
      </ContentTransition>
      <ContentTransition show={currentStep == 1}>
        <IntroStep1 />
      </ContentTransition>
      <ContentTransition show={currentStep == 2}>
        <IntroStep2 />
      </ContentTransition>
      <div className="flex justify-between mb-6">
        <div className="flex-1 flex">
          <ImmuneLink theme="GHOST-WHITE" href={NextJsRoutes.dashboardIndexPath}>
            Skip Intro
          </ImmuneLink>
        </div>
        <div className="flex-1 flex justify-center">
          <IntroStepIndicator currentStep={currentStep} maxSteps={3} setStep={setCurrentStep} />
        </div>
        <div className="flex-1 flex justify-end">
          {currentStep == 0 && (
            <Button theme="CTA" onClick={() => setCurrentStep(currentStep + 1)}>
              Start your mission
            </Button>
          )}
          {currentStep == 1 && (
            <Button theme="CTA" onClick={() => setCurrentStep(currentStep + 1)}>
              Got it
            </Button>
          )}
          {currentStep == 2 && (
            <ImmuneLink theme="SUCCESS" href={NextJsRoutes.dashboardIndexPath}>
              Let’s go
            </ImmuneLink>
          )}
        </div>
      </div>
    </>
  );
};
export default OnboardingIntro;
