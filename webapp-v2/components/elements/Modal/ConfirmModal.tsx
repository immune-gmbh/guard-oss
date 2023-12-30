import Button from 'components/elements/Button/Button';
import Modal from 'components/elements/Modal/Modal';
import React, { useState } from 'react';

interface IConfirmModal {
  headline?: string;
  text?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  TriggerComponent: React.FC<{ onClick: (e) => void }>;
  onConfirm: () => void;
}

const ConfirmModal: React.FC<IConfirmModal> = ({
  TriggerComponent,
  text,
  onConfirm,
  children,
  confirmLabel = 'Yes',
  cancelLabel = 'Cancel',
  headline = 'Are you sure?',
}) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <TriggerComponent
        onClick={(e) => {
          setIsOpen(true);
          e.preventDefault();
        }}
      />
      <Modal isOpen={isOpen} title={headline} closeModal={() => setIsOpen(false)}>
        <div className="space-y-2">{text}</div>
        {children && <div className="space-y-4">{children}</div>}
        <div className="mt-8 flex justify-between">
          <Button data-qa="no" onClick={() => setIsOpen(false)}>
            {cancelLabel}
          </Button>
          <Button
            theme="CTA"
            data-qa="yes"
            onClick={() => {
              onConfirm();
              setIsOpen(false);
            }}>
            {confirmLabel}
          </Button>
        </div>
      </Modal>
    </>
  );
};

export default ConfirmModal;
