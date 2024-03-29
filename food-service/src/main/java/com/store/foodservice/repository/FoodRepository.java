package com.store.foodservice.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;

import com.store.foodservice.model.Food;

public interface FoodRepository extends JpaRepository<Food, Long>{

}
