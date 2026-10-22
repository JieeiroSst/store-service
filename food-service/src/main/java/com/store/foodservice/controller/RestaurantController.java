package com.store.foodservice.controller;

import java.util.List;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.store.foodservice.model.Food;
import com.store.foodservice.model.Restaurant;
import com.store.foodservice.repository.FoodRepository;
import com.store.foodservice.repository.RestaurantRepository;

@RestController
@RequestMapping("/api/restaurants")
public class RestaurantController {
    private final RestaurantRepository restaurants;
    private final FoodRepository foods;

    public RestaurantController(RestaurantRepository restaurants, FoodRepository foods) {
        this.restaurants = restaurants;
        this.foods = foods;
    }

    @GetMapping("")
    public List<Restaurant> list(@RequestParam(required = false) String q) {
        List<Restaurant> all = q == null || q.isBlank() ? restaurants.findAll() : restaurants.findByNameContainingIgnoreCase(q.trim());
        return all.stream().filter(Restaurant::isActive).toList();
    }

    @GetMapping("/{id}")
    public ResponseEntity<Restaurant> get(@PathVariable("id") long id) {
        return restaurants.findById(id).map(ResponseEntity::ok).orElseGet(() -> ResponseEntity.notFound().build());
    }

    @GetMapping("/{id}/foods")
    public ResponseEntity<List<Food>> menu(@PathVariable("id") long id) {
        if (!restaurants.existsById(id)) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(foods.findByRestaurantId(id).stream().filter(Food::isActive).toList());
    }

    @PostMapping("")
    public ResponseEntity<Restaurant> create(@RequestBody Restaurant body) {
        if (body.getName() == null || body.getName().isBlank()) {
            return ResponseEntity.badRequest().build();
        }
        body.setId(0);
        return new ResponseEntity<>(restaurants.save(body), HttpStatus.CREATED);
    }

    @PutMapping("/{id}")
    public ResponseEntity<Restaurant> update(@PathVariable("id") long id, @RequestBody Restaurant body) {
        return restaurants.findById(id).map(existing -> {
            if (body.getName() != null) {
                existing.setName(body.getName());
            }
            existing.setDescription(body.getDescription());
            existing.setAddress(body.getAddress());
            existing.setImageUrl(body.getImageUrl());
            existing.setActive(body.isActive());
            existing.setCommissionRate(body.getCommissionRate());
            return ResponseEntity.ok(restaurants.save(existing));
        }).orElseGet(() -> ResponseEntity.notFound().build());
    }
}
