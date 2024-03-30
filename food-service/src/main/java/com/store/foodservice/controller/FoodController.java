package com.store.foodservice.controller;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.store.foodservice.usecase.FoodUsecase;
import com.store.foodservice.model.Food;

@CrossOrigin(origins = "http://localhost:8080")
@RestController
@RequestMapping("/api/food")
public class FoodController {
    @Autowired
    FoodUsecase foodUsecase;

    @SuppressWarnings("null")
    @PostMapping("")
    public ResponseEntity<Food> createFood(@RequestBody Food food) {
        try {
            Food _food = foodUsecase.save(food);
            return new ResponseEntity<>(_food, HttpStatus.CREATED);
        } catch (Exception e) {
            return new ResponseEntity<>(null, HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @PutMapping("/{id}")
    public ResponseEntity<Food> updateFood(@PathVariable("id") long id, @RequestBody Food food) {
        Optional<Food> FoodData = foodUsecase.findById(id);

        if (FoodData.isPresent()) {
            Food _food = FoodData.get();
            _food.setName(food.getName());
            _food.setDescription(food.getDescription());
            _food.setCategory(food.getCategory());
            _food.setPrice(food.getPrice());
            return new ResponseEntity<>(foodUsecase.update(_food), HttpStatus.OK);
        } else {
            return new ResponseEntity<>(HttpStatus.NOT_FOUND);
        }
    }

    @DeleteMapping("/")
    public ResponseEntity<HttpStatus> deleteAllFoods() {
        try {
            foodUsecase.deleteAll();
            return new ResponseEntity<>(HttpStatus.NO_CONTENT);
        } catch (Exception e) {
            return new ResponseEntity<>(HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<HttpStatus> deleteFood(@PathVariable("id") long id) {
        try {
            foodUsecase.deleteById(id);
            return new ResponseEntity<>(HttpStatus.NO_CONTENT);
        } catch (Exception e) {
            return new ResponseEntity<>(HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<Food> getFoodById(@PathVariable("id") long id) {
        Optional<Food> FoodData = foodUsecase.findById(id);

        if (FoodData.isPresent()) {
            return new ResponseEntity<>(FoodData.get(), HttpStatus.OK);
        } else {
            return new ResponseEntity<>(HttpStatus.NOT_FOUND);
        }
    }

    @GetMapping("/")
    public ResponseEntity<List<Food>> getAllFoods() {
        try {
            List<Food> Foods = new ArrayList<Food>();
            foodUsecase.findAll().forEach(Foods::add);
            if (Foods.isEmpty()) {
                return new ResponseEntity<>(HttpStatus.NO_CONTENT);
            }

            return new ResponseEntity<>(Foods, HttpStatus.OK);
        } catch (Exception e) {
            return new ResponseEntity<>(null, HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @GetMapping("/search")
    public ResponseEntity<List<Food>> getByNameFood(@RequestParam(required = true) String name) {
      try {
        List<Food> Foods = new ArrayList<Food>();

        foodUsecase.findByNameContaining(name).forEach(Foods::add);
  
        if (Foods.isEmpty()) {
          return new ResponseEntity<>(HttpStatus.NO_CONTENT);
        }
  
        return new ResponseEntity<>(Foods, HttpStatus.OK);
      } catch (Exception e) {
        return new ResponseEntity<>(null, HttpStatus.INTERNAL_SERVER_ERROR);
      }
    }
}
