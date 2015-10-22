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
    start_time = Time.now

    top_100_listings.each do |listing|
      Listing.create!(fields_for(listing))
    end

    puts "Scraped top 100 Alexa listings in %.2fs" % (Time.now - start_time)
  end

protected

  def self.top_100_listings
    Enumerator.new do |enum|
      TOP_100_URLS.each do |url|
        listings_on(url).each do |listing|
          enum.yield listing
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

    name = listing.css(".desc-paragraph a").text

    {
      description: listing.css(".description").text.strip, # there's a leading newline here
      global_rank: listing.css(".count").text.to_i,
      name:        name,
      url:        "http://" + name.downcase, # Alexa doesn't give the full URL, so add on http://
    }
  end

end

# ===== TASKS ======================================================================================

namespace :alexa do

  desc "Wipes and repopulates the `listings` table with the top 100 global sites from Alexa"
  task scrape_top_100: :environment do
    puts "Wiping `listings`"
    Listing.destroy_all # could use #delete_all here since there's no associations or callbacks

    AlexaScraper.scrape_top_100
  end

end
