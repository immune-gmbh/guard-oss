import AuthenticatedLink from 'components/elements/AuthenticatedLink/AuthenticatedLink';
import { format } from 'date-fns';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useInvoices } from 'hooks/invoices';
import getConfig from 'next/config';
import React from 'react';

interface IInvoicesTableProps {
  subscriptionId: string;
}

const InvoicesTable = ({ subscriptionId }: IInvoicesTableProps): JSX.Element => {
  const { publicRuntimeConfig } = getConfig();

  const { data: invoices } = useInvoices({ subscription_id: subscriptionId });

  return (
    <table className="w-full" role="table">
      <tbody>
        {invoices &&
          invoices
            .sort((n1, n2) => {
              const d1 = new Date(n1.finalizedAt);
              const d2 = new Date(n2.finalizedAt);

              if (d1 > d2) {
                return 1;
              }

              if (d1 < d2) {
                return -1;
              }

              return 0;
            })
            .map((invoice) => (
              <tr key={invoice.id} className="bg-white even:bg-gray-200 border-b">
                <td className="p-2">{format(new Date(invoice.finalizedAt), 'PPP')}</td>
                <td className="p-2">{invoice.total / 100.0}€</td>
                <td className="p-2">
                  <AuthenticatedLink
                    filename="Invoice"
                    qaLabel="invoice"
                    url={
                      publicRuntimeConfig.hosts.authSrv +
                      JsRoutesRails.download_v2_subscription_invoice_path({
                        subscription_id: subscriptionId,
                        id: invoice.id,
                      })
                    }>
                    Download PDF
                  </AuthenticatedLink>
                </td>
              </tr>
            ))}
      </tbody>
    </table>
  );
};
export default InvoicesTable;
