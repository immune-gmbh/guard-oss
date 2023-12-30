import classNames from 'classnames';
import TabBar from 'components/elements/TabBar/TabBar';
import TabItem from 'components/elements/TabBar/TabItem';
import { useState } from 'react';

interface ITabNavigation extends React.HTMLProps<HTMLDivElement> {
  bodyClassNames?: string;
  navigationPoints: {
    title: string;
    component: React.ReactNode;
  }[];
  activeTabIndex?: number;
}

export default function TabNavigation({
  navigationPoints,
  bodyClassNames,
  activeTabIndex = 0,
  ...rest
}: ITabNavigation): JSX.Element {
  const [currentTab, setCurrentTab] = useState(activeTabIndex);

  return (
    <div {...rest}>
      <TabBar>
        {navigationPoints.map((navigationPoint, index) => (
          <TabItem
            key={navigationPoint.title}
            title={navigationPoint.title}
            active={index === currentTab}
            onClick={() => setCurrentTab(index)}
          />
        ))}
      </TabBar>
      <div className={classNames(bodyClassNames)}>{navigationPoints[currentTab].component}</div>
    </div>
  );
}
