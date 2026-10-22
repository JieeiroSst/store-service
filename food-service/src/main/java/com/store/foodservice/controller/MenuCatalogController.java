package com.store.foodservice.controller;

import java.util.LinkedHashMap;
import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.store.foodservice.repository.FoodRepository;
import com.store.foodservice.repository.RestaurantRepository;

@RestController
@RequestMapping("/v1")
public class MenuCatalogController {
    private final RestaurantRepository restaurants;
    private final FoodRepository foods;

    public MenuCatalogController(RestaurantRepository restaurants, FoodRepository foods) {
        this.restaurants = restaurants;
        this.foods = foods;
    }

    @GetMapping("/restaurants/{id}")
    public ResponseEntity<Map<String, Object>> restaurant(@PathVariable("id") String id) {
        Long key = parseId(id);
        if (key == null) {
            return ResponseEntity.notFound().build();
        }
        return restaurants.findById(key).map(r -> {
            Map<String, Object> body = new LinkedHashMap<>();
            body.put("restaurant_id", String.valueOf(r.getId()));
            body.put("name", r.getName());
            body.put("is_active", r.isActive());
            body.put("commission_rate", r.getCommissionRate());
            return ResponseEntity.ok(body);
        }).orElseGet(() -> ResponseEntity.notFound().build());
    }

    @GetMapping("/menu-items/{id}")
    public ResponseEntity<Map<String, Object>> menuItem(@PathVariable("id") String id) {
        Long key = parseId(id);
        if (key == null) {
            return ResponseEntity.notFound().build();
        }
        return foods.findById(key).map(f -> {
            Map<String, Object> body = new LinkedHashMap<>();
            body.put("item_id", String.valueOf(f.getId()));
            body.put("restaurant_id", f.getRestaurantId() == null ? "" : String.valueOf(f.getRestaurantId()));
            body.put("name", f.getName());
            body.put("price", (double) f.getPrice());
            body.put("is_active", f.isActive() && f.getRestaurantId() != null);
            return ResponseEntity.ok(body);
        }).orElseGet(() -> ResponseEntity.notFound().build());
    }

    private static Long parseId(String raw) {
        try {
            long v = Long.parseLong(raw.trim());
            return v > 0 ? v : null;
        } catch (NumberFormatException e) {
            return null;
        }
    }
}
