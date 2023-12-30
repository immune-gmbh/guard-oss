import SubHeadline from 'components/elements/Headlines/SubHeadline';
import NavigationLink, {
  INavigationLinkProps,
} from 'components/elements/Navigations/NavigationLink';
import React from 'react';

interface INavigationSectionProps extends React.HTMLAttributes<HTMLDivElement> {
  headline?: string;
  menuItems: INavigationLinkProps[];
}

const NavigationSection: React.FC<INavigationSectionProps> = ({
  menuItems,
  headline,
  ...attributes
}) => {
  return (
    <nav {...attributes}>
      {headline ? <SubHeadline className="mb-4">{headline}</SubHeadline> : null}
      <ul className="text-white text-xl space-y-4">
        {menuItems.map((item) =>
          !item.hide ? <NavigationLink key={item.label} {...item} /> : null,
        )}
      </ul>
    </nav>
  );
};
export default NavigationSection;
