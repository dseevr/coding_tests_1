class ListingsController < ApplicationController

  before_action :set_listing, only: [:show, :edit, :update]

  def home
    redirect_to listings_path
  end

  def index
    # get rid of the pointless page=1
    if params[:page] == "1"
      redirect_to listings_path
    end

    @listings = Listing.paginate(page: params[:page], per_page: 25)
  end

  def update
    if @listing.update(listing_params)
      render json: {}, status: :ok
    else
      render json: @listing.errors, status: :unprocessable_entity
    end
  end

private

  def set_listing
    @listing = Listing.find(params[:id])
  end

  def listing_params
    params.require(:listing).permit(:name, :url, :description, :global_rank)
  end

end
