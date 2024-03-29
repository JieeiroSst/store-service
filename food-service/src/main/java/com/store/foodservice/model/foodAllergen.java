package com.store.foodservice.model;

import jakarta.persistence.*;

@Entity
@Table(name = "food_allergens")
public class FoodAllergen {
    @Id
    @Column(name = "food_id")
    private Long foodId;

    @Id
    @Column(name = "allergen_id")
    private Long allergenId;

    public Long getFoodId() {
        return foodId;
    }

    public void setFoodId(Long foodId) {
        this.foodId = foodId;
    }

    public Long getAllergenId() {
        return allergenId;
    }

    public void setAllergenId(Long allergenId) {
        this.allergenId = allergenId;
    }
}
