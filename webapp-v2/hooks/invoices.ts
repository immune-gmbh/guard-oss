import { JsRoutesRails } from 'generated/authsrvRoutes';
import useSWR from 'swr';
import { ApiResponse } from 'utils/api';
import { fetcher } from 'utils/fetcher';
import { SerializedInvoice } from 'utils/types';

export const useInvoices = ({ subscription_id }): ApiResponse<SerializedInvoice[]> => {
  const { data, error } = useSWR(
    JsRoutesRails.v2_subscription_invoices_path({ subscription_id }),
    fetcher,
  );

  return {
    data: data as SerializedInvoice[],
    isLoading: !error && !data,
    isError: error,
  };
};
