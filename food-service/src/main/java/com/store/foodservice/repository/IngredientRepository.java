package com.store.foodservice.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;

import com.store.foodservice.model.Ingredient;

public interface IngredientRepository  extends JpaRepository<Ingredient, Long>{

}
