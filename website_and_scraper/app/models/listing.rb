# NOTE: My Rails models are always divided into these sections, and I try to keep things in
#       alphabetical order and lined up where possible.  I also try not to exceed 100 columns
#       because I like to have multiple panes open at once in my editor.

class Listing < ActiveRecord::Base

  # ===== RELATIONSHIPS ============================================================================

  # ===== SCOPES ===================================================================================

  default_scope { order("global_rank ASC") }

  # ===== VALIDATIONS ==============================================================================

  # some Alexa listings have no description, so allow blanks.
  validates_presence_of :description, allow_blank: true

  validates_presence_of     :global_rank, null: false
  validates_numericality_of :global_rank, only_integer: true, greater_than_or_equal_to: 1

  # NOTE: We could also validate uniqueness of the other fields, too, but since we're scraping
  #       someone else's website, that's really their problem.  We're sorting on this field
  #       anyways in the default_scope above, so we'd need an index here anyways.  This validation
  #       is just for showing a friendly error instead of a Postgres unique index violation if a
  #       duplicate value is used.
  validates_uniqueness_of :global_rank

  validates_presence_of :name, allow_blank: false

  validates_presence_of :url, allow_blank: false

  # DISCLAIMER: Having worked in the advertising industry and on web stuff in general for a long
  #             time, validating URLs is a really strange thing.  You can use one of the crazy
  #             long regexes on the URI class, but you will most definitely receive user input
  #             at some point which, while technically invalid, DOES work in browsers and your
  #             users will get mad at you.  I experienced this firsthand at Triggit when fixing
  #             up their very poorly validated Rails code, and ultimately this is exactly the URL
  #             validation regex which was in place when I left.  Shocking, I know.
  #             It's also worthwhile to run String#strip in a before_* callback because of
  #             trailing whitespace which users will accidentally include.
  validates_format_of :url, with: /\Ahttps?:\/\//

  # ===== CALLBACKS ================================================================================

  # ===== CLASS METHODS ============================================================================

  # ===== INSTANCE METHODS =========================================================================

end
