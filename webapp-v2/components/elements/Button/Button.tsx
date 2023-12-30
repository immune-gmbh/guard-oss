import { ButtonClassNames } from 'components/elements/Button/ButtonStyles';

export interface ImmuneButtonProps extends React.HTMLProps<HTMLButtonElement> {
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
}

function Button({
  onClick,
  full,
  theme = 'MAIN',
  children,
  disabled,
  className,
  type,
  ...rest
}: ImmuneButtonProps): JSX.Element {
  return (
    <button
      onClick={onClick}
      type={type as 'button' | 'submit' | 'reset'}
      disabled={disabled}
      className={ButtonClassNames({ full, theme, disabled, className })}
      {...rest}>
      {children}
    </button>
  );
}

export default Button;
