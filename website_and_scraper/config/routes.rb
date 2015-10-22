Rails.application.routes.draw do

  resources :listings, only: [:edit, :index, :update]

  root to: "listings#home"

end
