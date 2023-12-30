import React from 'react';
import { SerializedUser } from 'utils/types';

interface IAvatarProps {
  user: Pick<SerializedUser, 'name' | 'email'>;
}

const getNameInitials = (name: string): string =>
  name
    .split(' ')
    .map((nameSegment) => nameSegment[0])
    .join('')
    .substring(0, 2);

const Avatar: React.FC<IAvatarProps> = ({ user }) => {
  return (
    <div className="rounded-full bg-red-cta w-10 h-10 flex justify-center items-center cursor-pointer text-white hover:bg-red-600">
      {getNameInitials(user.name)}
    </div>
  );
};
export default Avatar;
