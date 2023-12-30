
module V2
  class InvoicesController < V2::ApiBaseController
    require 'open-uri'
    load_and_authorize_resource :subscription
    load_and_authorize_resource :invoice, through: :subscription

    def index
      render json: InvoiceSerializer.new(@invoices).serializable_hash.to_json
    end

    def show
      render json: InvoiceSerializer.new(@invoice).serializable_hash.to_json
    end

    def download
      file = URI.open(@invoice.stripe_pdf_url)
      send_file(
        file.path,
        filename: @invoice.stripe_invoice_number,
        type: 'application/pdf',
        status: 200
      )
    end
  end
end
