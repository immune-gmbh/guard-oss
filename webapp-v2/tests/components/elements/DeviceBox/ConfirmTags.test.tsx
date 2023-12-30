/**
 * @jest-environment jsdom
 */
import { describe, expect, test } from '@jest/globals';
import { render, screen, fireEvent, within } from '@testing-library/react';
import ConfirmTags, {
  deviceTagsToUpdate,
  getCommonTags,
} from 'components/elements/DeviceBox/ConfirmTags';
import * as UseSearchTagsHook from 'hooks/tags';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { devicesMock, tagsMock } from 'tests/mocks';
import { ApiSrv } from 'types/apiSrv';

const mockUseSearchTags = jest.spyOn(UseSearchTagsHook, 'useSearchTags') as jest.Mock;

describe('confirm tags modal', () => {
  mockUseSearchTags.mockReturnValue({ tags: tagsMock, mutate: jest.fn() });

  test('it opens modal with common tags already selected', () => {
    render(
      <SelectedDevicesContext.Provider value={{ items: devicesMock, dispatch: () => ({}) }}>
        <ConfirmTags />
      </SelectedDevicesContext.Provider>,
    );

    const button = screen.getByText(/Set \/ delete tags/i);
    expect(button).toBeInstanceOf(HTMLButtonElement);

    fireEvent.click(button);

    const dialog = screen.getByRole('dialog');
    expect(dialog).toBeInTheDocument();

    const listedTags = within(dialog).queryAllByRole('button');

    // first tag is common one
    expect(listedTags[0]).not.toHaveStyle('border-color: transparent');
    // the other two are not selected
    expect(listedTags[1]).toHaveStyle('border-color: transparent');
    expect(listedTags[2]).toHaveStyle('border-color: transparent');
  });

  test('get common tags between selected devices', () => {
    const selectedDevices = [
      devicesMock[0],
      { ...devicesMock[1], tags: tagsMock.slice(0, 2) },
    ] as ApiSrv.Device[];

    const commonTags = getCommonTags(selectedDevices, tagsMock);

    expect(commonTags.length).toBe(2);
    expect(commonTags.map((tag) => tag.key)).toEqual(['Blah', 'Blub']);
  });

  describe('get correct tags to be updated', () => {
    test('add one tag to one, remove one from both', () => {
      const toBeUpdated = deviceTagsToUpdate(
        devicesMock,
        getCommonTags(devicesMock, tagsMock),
        tagsMock.slice(-2),
      );

      const device1770Tags = toBeUpdated[0].tags.map((tag) => tag.key);
      const device1772Tags = toBeUpdated[1].tags.map((tag) => tag.key);

      expect(device1770Tags).toEqual(['Blub', 'Foo']);
      expect(device1772Tags).toEqual(['Blub', 'Foo']);
    });

    test('adds newly created tags, removes the common one', () => {
      const toBeUpdated = deviceTagsToUpdate(devicesMock, getCommonTags(devicesMock, tagsMock), [
        { id: '', key: 'newTag1', score: 0.0 },
        { id: '', key: 'newTag2', score: 0.0 },
      ]);

      const device1770Tags = toBeUpdated[0].tags.map((tag) => tag.key);
      const device1772Tags = toBeUpdated[1].tags.map((tag) => tag.key);

      expect(device1770Tags).toEqual(['Blub', 'Foo', 'newTag1', 'newTag2']);
      expect(device1772Tags).toEqual(['newTag1', 'newTag2']);
    });

    test('add one tag and remove one to just one device', () => {
      const device1772 = devicesMock[1];
      const toBeUpdated = deviceTagsToUpdate([device1772], getCommonTags([device1772], tagsMock), [
        tagsMock[2],
      ]);

      expect(toBeUpdated[0].tags.map((tag) => tag.key)).toEqual(['Foo']);
    });
  });
});
