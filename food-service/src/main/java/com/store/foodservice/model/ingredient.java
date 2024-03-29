package com.store.foodservice.model;

import jakarta.persistence.*;

@Entity
@Table(name = "ingredient")
public class ingredient {
    @Id
	@GeneratedValue(strategy = GenerationType.AUTO)
	private long id;

    @Column(name = "name")
	private String name;

    @Column(name = "unit")
	private String unit;

    public long getId() {
        return id;
    }
    public String getName() {
        return name;
    }
    public String getUnit() {
        return unit;
    }

    public void setId(long id) {
        this.id = id;
    }
    public void setName(String name) {
        this.name = name;
    }
    public void setUnit(String unit) {
        this.unit = unit;
    }
}
