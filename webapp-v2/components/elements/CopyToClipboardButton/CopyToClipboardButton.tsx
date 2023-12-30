import { Transition } from '@headlessui/react';
import { CheckCircleIcon, ClipboardCopyIcon } from '@heroicons/react/outline';
import React, { useEffect, useState } from 'react';

const CHECKMARK_SHOWN_MILLIS = 2000;

interface ICopyToClipboardButton {
  text: string;
  height?: string;
}

export default function CopyToClipboardButton({
  text,
  height = '5',
}: ICopyToClipboardButton): JSX.Element {
  const [justCopied, setJustCopied] = useState(false);

  useEffect(() => {
    let timeout;
    if (justCopied) {
      timeout = setTimeout(() => {
        setJustCopied(false);
      }, CHECKMARK_SHOWN_MILLIS);
    }
    return () => {
      timeout && clearTimeout(timeout);
    };
  }, [justCopied]);
  const handleCopy = (): void => {
    navigator.clipboard.writeText(text);
    setJustCopied(true);
  };

  return (
    <div className={`h-${height}`}>
      <Transition
        show={justCopied}
        className="duration-200 transition ease-in-out"
        enter="delay-100"
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave=""
        leaveFrom="opacity-100"
        leaveTo="opacity-0">
        <CheckCircleIcon
          className={`absolute w-${height} h-${height} cursor-pointer text-green-notification hover:text-black`}
          onClick={handleCopy}
        />
      </Transition>
      <Transition
        show={!justCopied}
        className="duration-200 transition ease-in-out"
        enter="delay-100"
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave=""
        leaveFrom="opacity-100 "
        leaveTo="opacity-0">
        <ClipboardCopyIcon
          className={`absolute w-${height} h-${height} cursor-pointer text-gray-400 hover:text-black`}
          onClick={handleCopy}
        />
      </Transition>
    </div>
  );
}
