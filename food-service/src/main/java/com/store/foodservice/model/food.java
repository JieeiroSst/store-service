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
