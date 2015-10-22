require "open-uri"
require "nokogiri"

class AlexaScraper

  TOP_100_URLS = %w[
    http://www.alexa.com/topsites
    http://www.alexa.com/topsites/global;1
    http://www.alexa.com/topsites/global;2
    http://www.alexa.com/topsites/global;3
  ]

  def self.scrape_top_100
    top_100_listings.each do |entry|
      # TODO: store into PG
      puts entry
    end
  end

protected

  def self.top_100_listings
    Enumerator.new do |enum|
      TOP_100_URLS.each do |url|
        listings_on(url).each do |entry|
          enum.yield(fields_for(entry))
        end
      end
    end
  end

  def self.listings_on(url)
    puts "Fetching #{url}"

    page_data = open(url).read
    doc = Nokogiri::HTML(page_data)

    Enumerator.new do |enum|
      doc.css(".site-listing").each do |listing|
        enum.yield listing
      end
    end
  end

  def self.fields_for(listing)
    # remove the first span in the listing because it has the "More" and ellipsis text
    # looks like they misspelled "truncate" in their CSS
    listing.css("span.trucate").remove

    {
      description: listing.css(".description").text.strip, # there's a leading newline here
      domain:      listing.css(".desc-paragraph a").text,
      global_rank: listing.css(".count").text.to_i,
    }
  end

end

# ===== TASKS ======================================================================================

namespace :alexa do
  desc "Scrapes the top 100 global sites from Alexa and stuffs them into Postgres"
  task scrape_top_100: :environment do
    AlexaScraper.scrape_top_100
  end

end
