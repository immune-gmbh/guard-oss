import React from 'react';

interface IIntroStepIndicatorProps {
  currentStep: number;
  maxSteps: number;
  setStep: (step: number) => void;
}

const IntroStepIndicator: React.FC<IIntroStepIndicatorProps> = ({
  currentStep,
  maxSteps,
  setStep,
}) => {
  return (
    <div className="flex space-x-2 items-center ">
      {Array(maxSteps)
        .fill(0)
        .map((_value, i) => (
          <div key={i}>
            {i == currentStep && <div className="w-10 h-4 bg-red-cta rounded-full"></div>}
            {i != currentStep && (
              <div
                className="w-4 h-4 bg-white rounded-full border border-red-cta cursor-pointer"
                onClick={() => setStep(i)}></div>
            )}
          </div>
        ))}
    </div>
  );
};
export default IntroStepIndicator;
