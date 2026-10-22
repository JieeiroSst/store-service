package com.store.foodservice.model;

import jakarta.persistence.*;

@Entity
@Table(name = "foods")
public class Food {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private long id;

    @Column(name = "name")
    private String name;

    @Column(name = "description")
    private String description;

    @Column(name = "price")
    private long price;

    @Column(name = "category")
    private String category;

    @Column(name = "restaurant_id")
    private Long restaurantId;

    @Column(name = "is_active")
    private Boolean active;

    public Food() {
    }

    public Food(String name, String description, long price, String category) {
        this.name = name;
        this.description = description;
        this.price = price;
        this.category = category;
    }

    @com.fasterxml.jackson.annotation.JsonProperty("restaurant_id")
    public Long getRestaurantId() {
        return restaurantId;
    }

    @com.fasterxml.jackson.annotation.JsonProperty("restaurant_id")
    public void setRestaurantId(Long restaurantId) {
        this.restaurantId = restaurantId;
    }

    @com.fasterxml.jackson.annotation.JsonProperty("is_active")
    public boolean isActive() {
        return active == null || active;
    }

    @com.fasterxml.jackson.annotation.JsonProperty("is_active")
    public void setActive(Boolean active) {
        this.active = active;
    }

    public String getCategory() {
        return category;
    }

    public String getDescription() {
        return description;
    }

    public long getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public long getPrice() {
        return price;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public void setId(long id) {
        this.id = id;
    }

    public void setName(String name) {
        this.name = name;
    }

    public void setPrice(long price) {
        this.price = price;
    }
}
