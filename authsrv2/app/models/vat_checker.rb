class VatChecker

  Countries = [
      'Austria',
      'Belgium',
      'Bulgaria',
      'Croatia',
      'Cyprus',
      'Czech Republic',
      'Denmark',
      'Estonia',
      'Finland',
      'France',
      'Germany',
      'Greece',
      'Hungary',
      'Ireland',
      'Italy',
      'Latvia',
      'Lithuania',
      'Luxembourg',
      'Malta',
      'Netherlands',
      'Poland',
      'Portugal',
      'Romania',
      'Slovakia',
      'Slovenia',
      'Spain',
      'Sweden',
      'United Kingdom of Great Britain and Northern Ireland'
  ].freeze

  StripeTaxIDTypes = {
    'Australia' => {
      11 => 'au_abn', #	Australian Business Number (AU ABN)	12345678912
      12 => 'au_arn', #	Australian Taxation Office Reference Number	123456789123
    },
    'Austria' => 'eu_vat', #	European VAT number	ATU12345678
    'Belgium' => 'eu_vat', #	European VAT number	BE0123456789
    'Brazil' => {
      18 => 'br_cnpj',#	Brazilian CNPJ number	01.234.456/5432-10
      14 => 'br_cpf', #	Brazilian CPF number	123.456.789-87
    } ,
    'Bulgaria' => 'eu_vat', #	European VAT number	BG0123456789
    'Canada' => {
      9 => 'ca_bn', #	Canadian BN	123456789
      15 => 'ca_gst_hst', #	Canadian GST/HST number	123456789RT0002
      13 => 'ca_pst_bc', #	Canadian PST number (British Columbia)	PST-1234-5678
      8 => 'ca_pst_mb', #	Canadian PST number (Manitoba)	123456-7
      7 => 'ca_pst_sk', #	Canadian PST number (Saskatchewan)	1234567
      16 => 'ca_qst', #	Canadian QST number (Québec)	1234567890TQ1234
    },
    'Chile' => 'cl_tin', #	Chilean TIN	12.345.678-K
    'Croatia' => 'eu_vat', #	European VAT number	HR12345678912
    'Cyprus' => 'eu_vat', #	European VAT number	CY12345678Z
    'Czech Republic' => 'eu_vat', #	European VAT number	CZ1234567890
    'Denmark' => 'eu_vat', #	European VAT number	DK12345678
    'Estonia' => 'eu_vat', #	European VAT number	EE123456789
    'Finland' => 'eu_vat', #	European VAT number	FI12345678
    'France' => 'eu_vat', #	European VAT number	FRAB123456789
    'Georgia' => 'ge_vat', #	Georgian VAT	123456789
    'Germany' => 'eu_vat', #	European VAT number	DE123456789
    'Greece' => 'eu_vat', #	European VAT number	EL123456789
    'Hong Kong' => 'hk_br', #	Hong Kong BR number	12345678
    'Hungary' => 'eu_vat', #	European VAT number	HU12345678912
    'Iceland' => 'is_vat', #	Icelandic VAT	123456
    'India' => 'in_gst', #	Indian GST number	12ABCDE3456FGZH
    'Indonesia' => 'id_npwp', #	Indonesian NPWP number	12.345.678.9-012.345
    'Ireland' => 'eu_vat', #	European VAT number	IE1234567AB
    'Israel' => 'il_vat', #	Israel VAT	000012345
    'Italy' => 'eu_vat', #	European VAT number	IT12345678912
    'Japan' => {
      13 => 'jp_cn', #	Japanese Corporate Number (*Hōjin Bangō*)	1234567891234
      5 => 'jp_rn', #	Japanese Registered Foreign Businesses' Registration Number (*Tōroku Kokugai Jigyōsha no Tōroku Bangō*)	12345
    },
    'Latvia' => 'eu_vat', #	European VAT number	LV12345678912
    'Liechtenstein' => 'li_uid', #	Liechtensteinian UID number	CHE123456789
    'Lithuania' => 'eu_vat', #	European VAT number	LT123456789123
    'Luxembourg' => 'eu_vat', #	European VAT number	LU12345678
    'Malaysia' => {
      8 => 'my_frp', #	Malaysian FRP number	12345678
      10 => 'my_itn', #	Malaysian ITN	C 1234567890
      17 => 'my_sst', #	Malaysian SST number	A12-3456-78912345
    },
    'Malta' => 'eu_vat', #	European VAT number	MT12345678
    'Mexico' => 'mx_rfc', #	Mexican RFC number	ABC010203AB9
    'Netherlands' => 'eu_vat', #	European VAT number	NL123456789B12
    'New Zealand' => 'nz_gst', #	New Zealand GST number	123456789
    'Norway' => 'no_vat', #	Norwegian VAT number	123456789MVA
    'Poland' => 'eu_vat', #	European VAT number	PL1234567890
    'Portugal' => 'eu_vat', #	European VAT number	PT123456789
    'Romania' => 'eu_vat', #	European VAT number	RO1234567891
    'Russia' => {
      10 => 'ru_inn', #	Russian INN	1234567891
      9 => 'ru_kpp', #	Russian KPP	123456789
    },
    'Saudi Arabia' => 'sa_vat', #	Saudi Arabia VAT	123456789012345
    'Singapore' => {
      /^M.+$/ => 'sg_gst', #	Singaporean GST	M12345678X
      /^[0-9].+$/=> 'sg_uen', #	Singaporean UEN	123456789F
    },
    'Slovakia' => 'eu_vat', #	European VAT number	SK1234567891
    'Slovenia' => 'eu_vat', #	European VAT number	SI12345678
    'South Africa' => 'za_vat', #	South African VAT number	4123456789
    'South Korea' => 'kr_brn', #	Korean BRN	123-45-67890
    'Spain' => {
      9 => 'es_cif', #	Spanish CIF number	A12345678
      11 => 'eu_vat', #	European VAT number	ESA1234567Z
    },
    'Sweden' => 'eu_vat', #	European VAT number	SE123456789123
    'Switzerland' => 'ch_vat', #	Switzerland VAT number	CHE-123.456.789 MWST
    'Taiwan' => 'tw_vat', #	Taiwanese VAT	12345678
    'Thailand' => 'th_vat', #	Thai VAT	1234567891234
    'Ukraine' => 'ua_vat', #	Ukrainian VAT	123456789
    'United Arab Emirates' => 'ae_trn', #	United Arab Emirates TRN	123456789012345
    'United Kingdom' => {
      /^GB.+$/ => 'gb_vat', #	United Kingdom VAT number	GB123456789
      /^XI.+$/ => 'eu_vat', #	Northern Ireland VAT number	XI123456789
    },
    'United States' => 'us_ein', #	United States EIN	12-3456789
  }

  def self.apply_vat?(country)
    Countries.include?(country)
  end

  def self.valid_vat_id_for_country(vat_id, country)
    vat_country = Valvat.new(vat_id).exists?(detail: true)
    vat_country && vat_country[:country_code] == ISO3166::Country.find_country_by_name(country).alpha2
  end

  def self.stripe_taxid_for_country(country, vat_id)
    case (candidates = StripeTaxIDTypes[country])
    when String
      candidates
    when Hash
      type = candidates.select do |pat, ty| 
        case pat
          when Integer
            vat_id.size == pat
          when Regexp
            pat.match? vat_id
          else
            false
          end
      end.values.first
      raise ArgumentError.new 'unrecognized VAT ID format' if !type
      type
    else
      raise ArgumentError.new 'unrecognized VAT ID format'
    end
  end
end
