package com.store.foodservice.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;

import com.store.foodservice.model.Restaurant;

public interface RestaurantRepository extends JpaRepository<Restaurant, Long> {
    List<Restaurant> findByNameContainingIgnoreCase(String name);
}
