class InvoiceSerializer < BaseSerializer
  attributes :total, :status, :stripe_invoice_number, :finalized_at
end
