package com.store.foodservice.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;

import com.store.foodservice.model.FoodAllergen;

public interface FoodAllergenRepository  extends JpaRepository<FoodAllergen, Long>{

}
