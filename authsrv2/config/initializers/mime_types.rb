# Be sure to restart your server when you modify this file.

# Add new mime types for use in respond_to blocks:
# Mime::Type.register "text/richtext", :rtf

json_mime_types = %W(
  application/vnd.api+json
  text/x-json
  application/json
)
Mime::Type.unregister :json
Mime::Type.register 'application/json', :json, json_mime_types
