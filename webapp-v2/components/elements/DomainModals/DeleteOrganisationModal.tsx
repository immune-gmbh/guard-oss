import { ExclamationIcon, TrashIcon } from '@heroicons/react/solid';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import Modal from 'components/elements/Modal/Modal';
import React, { useState } from 'react';
import { SerializedOrganisation } from 'utils/types';

interface IDeleteOrganisationModal {
  organisation: Pick<SerializedOrganisation, 'name' | 'id'>;
  onDelete: (organisation) => void;
}

const DeleteOrganisationModal: React.FC<IDeleteOrganisationModal> = ({
  organisation,
  onDelete,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [confirmDeleteText, setConfirmDeleteText] = useState('');

  const onConfirm = (): void => {
    setIsOpen(false);
    onDelete(organisation);
  };

  const parameterizedOrganisationName = organisation.name;

  return (
    <>
      <TrashIcon
        className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
        data-qa="org-delete"
        onClick={() => setIsOpen(true)}
      />
      <Modal
        isOpen={isOpen}
        title="Do your really want to delete the following organisation?"
        closeModal={() => setIsOpen(false)}>
        <div className="space-y-2">
          <Headline size={4}>{organisation.name}</Headline>
        </div>
        <p className="text-red-notification mt-6 mb-6">
          <ExclamationIcon className="inline h-4 mr-2" />
          Deleting an organisation can’t be undone!
        </p>
        <p>
          Please type <span className="font-bold">{parameterizedOrganisationName}</span> to confirm
        </p>
        <Input
          qaLabel="confirm-del"
          onChangeValue={setConfirmDeleteText}
          value={confirmDeleteText}
          aria-label="confirm-del"
        />
        <div className="mt-4 flex justify-between">
          <Button onClick={() => setIsOpen(false)}>Cancel</Button>
          <Button
            type="CTA"
            onClick={onConfirm}
            disabled={confirmDeleteText != parameterizedOrganisationName}>
            Delete
          </Button>
        </div>
      </Modal>
    </>
  );
};

export default DeleteOrganisationModal;
