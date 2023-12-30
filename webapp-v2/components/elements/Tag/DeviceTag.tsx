import XIcon from '@heroicons/react/solid/XIcon';
import { ApiSrv } from 'types/apiSrv';

function componentToHex(c: number): string {
  const hex = c.toString(16);
  return hex.length == 1 ? '0' + hex : hex;
}

function rgbToHex(r: number, g: number, b: number): string {
  return '#' + componentToHex(r) + componentToHex(g) + componentToHex(b);
}

function pastelColour(inputStr: string): { red: number; green: number; blue: number } {
  //TODO: adjust base colour values below based on theme
  const baseRed = 128;
  const baseGreen = 128;
  const baseBlue = 128;

  //lazy seeded random hack to get values from 0 - 256
  //for seed just take bitwise XOR of first two chars
  let seed = inputStr.charCodeAt(0) ^ inputStr.charCodeAt(1) ^ inputStr.charCodeAt(2);
  const rand_1 = Math.abs(Math.sin(seed++) * 10000) % 256;
  const rand_2 = Math.abs(Math.sin(seed++) * 10000) % 256;
  const rand_3 = Math.abs(Math.sin(seed++) * 10000) % 256;

  //build colour
  const red = Math.round((rand_1 + baseRed) / 2);
  const green = Math.round((rand_2 + baseGreen) / 2);
  const blue = Math.round((rand_3 + baseBlue) / 2);

  return { red, green, blue };
}

interface IDeviceTag {
  tag: ApiSrv.Tag;
  selected?: boolean;
  onClick?: () => void;
  remove?: boolean;
}

export default function DeviceTag({ tag, selected, onClick, remove }: IDeviceTag): JSX.Element {
  const tagName = typeof tag === 'string' ? tag : tag?.key;

  const { red, green, blue } = pastelColour(tagName);
  const bgColor = `rgba(${red}, ${green}, ${blue}, 0.1)`;
  const textColor = rgbToHex(red, green, blue);

  return (
    <div
      role="button"
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        onClick && onClick();
      }}
      style={{
        backgroundColor: bgColor,
        color: textColor,
        borderColor: selected ? textColor : 'transparent',
      }}
      className="flex gap-2.5 rounded-full border py-1 px-2.5 items-center whitespace-nowrap text-sm font-bold cursor-pointer hover:opacity-60">
      {tagName}
      {remove && <XIcon className="h-3 w-3" />}
    </div>
  );
}
