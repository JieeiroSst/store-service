package com.store.foodservice.model;

import jakarta.persistence.*;

@Entity
@Table(name = "allergens")
public class Allergens {
    @Id
	@GeneratedValue(strategy = GenerationType.AUTO)
	private long id;

    @Column(name = "name")
	private String name;

    public Allergens(String name) {
        this.name = name;
    }

    public long getId() {
        return id;
    }
    public String getName() {
        return name;
    }
    public void setId(long id) {
        this.id = id;
    }
    public void setName(String name) {
        this.name = name;
    }
}
