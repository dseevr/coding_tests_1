class CreateListings < ActiveRecord::Migration
  def change
    create_table :listings do |t|
      t.string :name,         null: false
      t.string :url,          null: false
      t.string :description,  null: false
      t.integer :global_rank, null: false

      t.timestamps null: false
    end

    add_index :listings, :global_rank, unique: true
  end
end
