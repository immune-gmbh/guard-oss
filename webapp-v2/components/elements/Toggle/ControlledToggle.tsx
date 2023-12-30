import Toggle, { IToggleProps } from 'components/elements/Toggle/Toggle';
import { Controller, ControllerProps } from 'react-hook-form';

interface IControlledToggleProps
  extends Omit<ControllerProps, 'render'>,
    Omit<IToggleProps, 'checked' | 'onChange'> {}

const ControlledToggle: React.FC<IControlledToggleProps> = ({ control, name, ...rest }) => {
  return (
    <Controller
      name={name}
      control={control}
      render={({ field }) => <Toggle {...rest} checked={field.value} onChange={field.onChange} />}
    />
  );
};
export default ControlledToggle;
