import Button from 'components/elements/Button/Button';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import DeviceTag from 'components/elements/Tag/DeviceTag';
import { useChangeTagsOfDevice, USE_DEVICES_URL } from 'hooks/devices';
import { useSearchTags } from 'hooks/tags';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { useContext, useEffect, useRef, useState } from 'react';
import { toast } from 'react-toastify';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { mutate } from 'swr';
import { ApiSrv } from 'types/apiSrv';

export const deviceTagsToUpdate = (
  selectedDevices: ApiSrv.Device[],
  initCommonTags: Array<any>,
  selectedTags: ApiSrv.Tag[],
) => {
  return selectedDevices.map((device) => {
    const nonCommonPreviousTags = (device.tags || []).filter(
      (tag) => !initCommonTags.some((cTag) => cTag.id == tag.id),
    );
    // avoid duplicates
    const selectedTagsNoPrevious = selectedTags.filter(
      (tag) => !nonCommonPreviousTags.some((cTag) => cTag.id == tag.id),
    );
    return {
      id: device.id,
      tags: [...nonCommonPreviousTags, ...selectedTagsNoPrevious],
    };
  });
};

export const getCommonTags = (selectedDevices: ApiSrv.Device[], availableTags: ApiSrv.Tag[]) => {
  const deviceIdsPerTag = selectedDevices.reduce((itemTags, currItem) => {
    (currItem.tags || []).forEach((tag) => {
      itemTags[tag.id] = itemTags[tag.id] || [];
      itemTags[tag.id].push(currItem.id);
    });
    return itemTags;
  }, {});

  return availableTags.filter((tag) => {
    if (!(tag.id in deviceIdsPerTag)) return false;

    return deviceIdsPerTag[tag.id].length == selectedDevices.length;
  });
};

export default function ConfirmTags(): JSX.Element {
  const [newTag, setNewTag] = useState('');
  const [availableTags, setAvailableTags] = useState<ApiSrv.Tag[]>([]);
  const [selectedTags, setSelectedTags] = useState<ApiSrv.Tag[]>([]);
  const selectedDevices = useContext(SelectedDevicesContext);
  const { tags, mutate: mutateTags } = useSearchTags({ query: '' });
  const initCommonTags = useRef([]);

  useEffect(() => {
    if (availableTags.length === 0 && tags.length > 0) {
      setAvailableTags([...tags]);
    }
  }, [tags]);

  useEffect(() => {
    initCommonTags.current = getCommonTags(selectedDevices.items, availableTags);
    setSelectedTags(initCommonTags.current);
  }, [availableTags, selectedDevices]);

  const tagDevice = useChangeTagsOfDevice();

  const toggleTag = (tag: ApiSrv.Tag): void => {
    const newTags = selectedTags.map((t) => t.id).includes(tag.id)
      ? selectedTags.filter((t) => t.id !== tag.id)
      : [...selectedTags, tag];

    setSelectedTags(newTags);
  };

  const newTagTemplate = (name: string): ApiSrv.Tag => ({
    id: '',
    key: name,
    score: 0.0,
  });

  const updateTags = () => {
    const toMutate = deviceTagsToUpdate(
      selectedDevices.items,
      initCommonTags.current,
      selectedTags,
    );

    return toMutate.map((deviceTags) => {
      tagDevice.mutate(deviceTags);
    });
  };

  return (
    <ConfirmModal
      headline="Set Tags"
      confirmLabel={`Set ${selectedTags.length} Tags for ${selectedDevices.items.length} devices`}
      onConfirm={() => {
        Promise.all(updateTags()).then(() => {
          mutate(USE_DEVICES_URL)
            .then(() => {
              toast.success(`Set tags for ${selectedDevices.items.length} devices`);
              selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
              mutateTags();
            })
            .then(() => {
              location.reload();
            });
        });
      }}
      TriggerComponent={(props) => (
        <Button theme="CTA" full={true} {...props} disabled={selectedDevices.items.length == 0}>
          Set / delete tags
        </Button>
      )}>
      <div className="-mt-4">
        <b>{selectedDevices.items.length}</b> Devices selected
        <div className="mt-8 ">Select multiple tags or create new ones</div>
        <ul className="flex flex-wrap gap-2 my-5">
          {availableTags.map((tag) => (
            <li key={tag.id}>
              <DeviceTag
                tag={tag}
                selected={selectedTags.map((t) => t.id).includes(tag.id)}
                onClick={() => toggleTag(tag)}
              />
            </li>
          ))}
        </ul>
        <div className="flex space-y-4 flex-col text-base">
          <span>Create a new Tag</span>
          <input
            type="text"
            placeholder="Press Enter to create your desired tag"
            className="w-1/2 text-base rounded text-primary"
            value={newTag}
            onChange={(e) => setNewTag(e.currentTarget.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                setAvailableTags([...availableTags, newTagTemplate(newTag)]);
                setSelectedTags([...selectedTags, newTagTemplate(newTag)]);
                setNewTag('');
              }
            }}
          />
          <span>Press enter to create your desired tag</span>
        </div>
      </div>
    </ConfirmModal>
  );
}
