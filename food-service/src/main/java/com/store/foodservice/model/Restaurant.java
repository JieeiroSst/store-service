package com.store.foodservice.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.persistence.*;

@Entity
@Table(name = "restaurants")
public class Restaurant {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private long id;

    @Column(name = "name", nullable = false)
    private String name;

    @Column(name = "description")
    private String description;

    @Column(name = "address")
    private String address;

    @Column(name = "image_url")
    private String imageUrl;

    @Column(name = "is_active")
    private Boolean active;

    @Column(name = "commission_rate")
    private Double commissionRate;

    public Restaurant() {
    }

    public Restaurant(String name, String address) {
        this.name = name;
        this.address = address;
    }

    public long getId() {
        return id;
    }

    public void setId(long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getAddress() {
        return address;
    }

    public void setAddress(String address) {
        this.address = address;
    }

    @JsonProperty("image_url")
    public String getImageUrl() {
        return imageUrl;
    }

    @JsonProperty("image_url")
    public void setImageUrl(String imageUrl) {
        this.imageUrl = imageUrl;
    }

    @JsonProperty("is_active")
    public boolean isActive() {
        return active == null || active;
    }

    @JsonProperty("is_active")
    public void setActive(Boolean active) {
        this.active = active;
    }

    @JsonProperty("commission_rate")
    public double getCommissionRate() {
        return commissionRate == null ? 0 : commissionRate;
    }

    @JsonProperty("commission_rate")
    public void setCommissionRate(Double commissionRate) {
        this.commissionRate = commissionRate;
    }
}
