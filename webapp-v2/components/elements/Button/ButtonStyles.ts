import classNames from 'classnames';

interface ButtonClassNameProps {
  theme?:
    | 'MAIN'
    | 'SECONDARY'
    | 'SECONDARY-RED'
    | 'CTA'
    | 'GHOST-WHITE'
    | 'GHOST-RED'
    | 'SUCCESS'
    | 'WHITE';
  full?: boolean;
  disabled?: boolean;
  className?: string;
}

export function ButtonClassNames({
  full,
  theme,
  disabled,
  className,
}: ButtonClassNameProps): string {
  return classNames(
    'immune-button',
    'flex items-center text-center justify-center font-bold border-2 rounded-m active:border-dashed',
    {
      'py-2': !className?.match(/py-[0-9]{1,2}/gm),
      'px-7': !className?.match(/px-[0-9]{1,2}/gm),
    },
    {
      [className]: !!className,
      'w-full': full,
      'border-purple-500': theme == 'MAIN' || theme === 'SECONDARY',
      'bg-purple-500 text-white': theme === 'MAIN',
      'border-red-cta bg-red-cta text-white': theme === 'CTA',
      'hover:bg-purple-200': theme === 'MAIN' && !disabled,
      'hover:border-purple-200': theme !== 'CTA' && !disabled,
      'text-white bg-green-notification border-transparent focus:border-green-200 hover:border-green-200':
        theme === 'SUCCESS',
      'opacity-60': disabled,
      'border-red-cta text-white': theme === 'SECONDARY-RED',
      'text-purple-500 bg-white': theme === 'WHITE',
    },
  );
}
