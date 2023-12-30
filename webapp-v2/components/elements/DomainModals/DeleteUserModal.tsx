import { TrashIcon } from '@heroicons/react/solid';
import Button from 'components/elements/Button/Button';
import Modal from 'components/elements/Modal/Modal';
import React, { useState } from 'react';

interface IDeleteUserModal {
  user;
  onDelete: (user) => void;
}

const DeleteUserModal: React.FC<IDeleteUserModal> = ({ user, onDelete }) => {
  const [isOpen, setIsOpen] = useState(false);

  const onConfirm = (): void => {
    setIsOpen(false);
    onDelete(user);
  };

  return (
    <>
      <TrashIcon
        className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
        data-qa="user-del"
        onClick={() => setIsOpen(true)}
      />
      <Modal
        isOpen={isOpen}
        title="Do your really want to delete the following user?"
        closeModal={() => setIsOpen(false)}>
        <div className="space-y-2">
          <p>{user.name}</p>
          <p>{user.email}</p>
          <p>{user.role}</p>
        </div>
        <div className="mt-4 flex justify-between">
          <Button onClick={() => setIsOpen(false)}>Reject</Button>
          <Button theme="CTA" onClick={onConfirm}>
            Delete
          </Button>
        </div>
      </Modal>
    </>
  );
};

export default DeleteUserModal;
