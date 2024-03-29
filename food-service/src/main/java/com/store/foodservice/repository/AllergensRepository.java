package com.store.foodservice.repository;

import org.springframework.data.jpa.repository.JpaRepository;

import com.store.foodservice.model.Allergens;

public interface AllergensRepository extends JpaRepository<Allergens, Long> {

}
