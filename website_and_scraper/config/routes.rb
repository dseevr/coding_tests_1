Rails.application.routes.draw do

  resources :listings, only: [:index, :update]

  root to: "listings#home"

end
