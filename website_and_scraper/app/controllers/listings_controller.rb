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

  def show
  end

  def edit
  end

  def update
    respond_to do |format|
      if @listing.update(listing_params)
        format.html { redirect_to @listing, notice: 'Listing was successfully updated.' }
        format.json { render :show, status: :ok, location: @listing }
      else
        format.html { render :edit }
        format.json { render json: @listing.errors, status: :unprocessable_entity }
      end
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
