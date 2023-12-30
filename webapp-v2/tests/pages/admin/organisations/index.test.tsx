/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import * as OrganisationsHook from 'hooks/organisations';
import { useRouter } from 'next/router';
import AdminOrganisations from 'pages/admin/organisations';
import { serializedOrganisation, serializedOrganisations } from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseOrganisations = jest.spyOn(OrganisationsHook, 'useOrganisations');
const mockUpdateOrganisation = jest.spyOn(OrganisationsHook, 'useUpdateOrganisation');

describe('organisations/index', () => {
  let mutateOrganisation: jest.Mock<() => Promise<unknown>>;

  const getTableElements = () => {
    const orgsTable = screen.queryByRole('table');
    const orgsTableBody = within(orgsTable).queryAllByRole('rowgroup')[1]; //tbody

    return {
      orgsTable,
      orgsTableBody,
      orgRows: within(orgsTableBody).queryAllByRole('row'),
    };
  };

  beforeEach(() => {
    mutateOrganisation = jest.fn(() => Promise.resolve(serializedOrganisation));

    mockUseRouter.mockReturnValue({
      pathname: '/organisations',
    });

    mockUseOrganisations.mockImplementation(() => ({
      data: serializedOrganisations,
      isLoading: false,
      isError: undefined,
    }));

    mockUpdateOrganisation.mockImplementation(() => ({
      mutate: mutateOrganisation,
      data: serializedOrganisation,
      isError: false,
      isLoading: false,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('renders the component correctly, when still loading', async () => {
    mockUseOrganisations.mockImplementationOnce(() => ({
      data: [],
      isLoading: true,
      isError: undefined,
    }));

    render(<AdminOrganisations />);

    expect(screen.queryByText('Organisation Management')).toBeInTheDocument();
    expect(screen.queryByRole('status')).toBeInTheDocument();
  });

  test('renders the component correctly, when no loading anymore', async () => {
    render(<AdminOrganisations />);

    const { orgRows } = getTableElements();

    expect(screen.queryByText('Organisation Management')).toBeInTheDocument();

    expect(orgRows.length).toBe(3);
    expect(within(orgRows[0]).getByText("Kai's Organisation")).toBeInTheDocument();
    expect(within(orgRows[1]).getByText("Testy's Organisation")).toBeInTheDocument();
    expect(within(orgRows[2]).getByText('Payment reminder Organisation')).toBeInTheDocument();
  });

  test('can update organisations name on table', async () => {
    let orgRowFirst: HTMLElement;

    render(<AdminOrganisations />);

    await waitFor(() => {
      const { orgRows } = getTableElements();
      orgRowFirst = orgRows[0]
    })

    expect(within(orgRowFirst).getByText("Kai's Organisation")).toBeInTheDocument();

    const firstOrgCells = within(orgRowFirst).queryAllByRole('cell');
    const firstOrgOptionsCell = firstOrgCells.slice(-1)[0];

    const options = firstOrgOptionsCell.firstElementChild?.children;
    expect(options.length).toBe(4);

    const editOption = options.item(3);
    fireEvent.click(editOption);

    // options are now check or cancel
    const confirmOptions = firstOrgOptionsCell.firstElementChild?.children;
    expect(confirmOptions.length).toBe(2);

    const nameInput = firstOrgCells[0].firstElementChild;
    fireEvent.change(nameInput, { target: { value: 'New Organisation' } });
    fireEvent.blur(nameInput);

    await act(async () => {
      fireEvent.click(confirmOptions[0]);

      await waitFor(() => {
        expect(mutateOrganisation).toHaveBeenCalled();
        expect(within(orgRowFirst).queryByText('New Organisation')).toBeInTheDocument();
      });
    });
  });

  test('table shows initial name if changes discarded', async () => {
    let orgRowFirst: HTMLElement;

    render(<AdminOrganisations />);

    await waitFor(() => {
      const { orgRows } = getTableElements();
      orgRowFirst = orgRows[0]
    })

    expect(within(orgRowFirst).getByText("Kai's Organisation")).toBeInTheDocument();

    const firstOrgCells = within(orgRowFirst).queryAllByRole('cell');
    const firstOrgOptionsCell = firstOrgCells.slice(-1)[0];

    // screen.debug(firstOrgOptionsCell, Infinity)
    const options = firstOrgOptionsCell.firstElementChild?.children;
    expect(options.length).toBe(4);

    const editOption = options.item(3);
    fireEvent.click(editOption);

    // options are now check or cancel
    const confirmOptions = firstOrgOptionsCell.firstElementChild?.children;
    expect(confirmOptions.length).toBe(2);

    const nameInput = firstOrgCells[0].firstElementChild;
    fireEvent.change(nameInput, { target: { value: 'New Organisation' } });
    fireEvent.blur(nameInput);

    await act(async () => {
      fireEvent.click(confirmOptions[1]);

      await waitFor(() => {
        expect(mutateOrganisation).not.toHaveBeenCalled();
        expect(within(orgRowFirst).queryByText('New Organisation')).not.toBeInTheDocument();
        expect(within(orgRowFirst).getByText("Kai's Organisation")).toBeInTheDocument();
      });
    });
  });
});
