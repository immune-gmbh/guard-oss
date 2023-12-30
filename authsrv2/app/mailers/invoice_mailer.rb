class InvoiceMailer < ApplicationMailer
  def invoice_paid_email(invoice)
    # TODO: check which users should be eligible for invoices in general
    # For now only :owner receive them if they want to
    invoice.subscription.organisation.memberships.role_owner.each do |membership|
      next unless membership.notify_invoice

      @user = membership.user

      mail(to: @user.email, subject: "Invoice #{invoice.stripe_invoice_number} for #{invoice.subscription.organisation.name}")
    end
  end

  def payment_failed_email(invoice)
    invoice.subscription.organisation.memberships.role_owner.each do |membership|
      next unless membership.notify_invoice

      @user = membership.user
      @organisation = invoice.subscription.organisation

      mail(to: @user.email, subject: "Payment failed for #{@organisation.name}")
    end
  end
end
