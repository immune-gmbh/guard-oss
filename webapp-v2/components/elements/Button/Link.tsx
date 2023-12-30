import { ButtonClassNames } from 'components/elements/Button/ButtonStyles';
import NextLink from 'next/link';
import { UrlObject } from 'url';

interface ImmuneLinkProps
  extends Pick<
    React.HTMLProps<HTMLAnchorElement>,
    'className' | 'disabled' | 'children' | 'target'
  > {
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
  isButton?: boolean;
  href: string | UrlObject;
}

function Link({
  full,
  theme = 'MAIN',
  children,
  disabled,
  className,
  href,
  isButton = true,
  ...rest
}: ImmuneLinkProps): JSX.Element {
  return (
    <NextLink href={href} passHref>
      <a
        className={isButton ? ButtonClassNames({ full, theme, disabled, className }) : className}
        {...rest}>
        {children}
      </a>
    </NextLink>
  );
}

export default Link;
