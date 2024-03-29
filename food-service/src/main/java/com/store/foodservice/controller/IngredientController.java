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

import com.store.foodservice.model.Ingredient;
import com.store.foodservice.repository.IngredientRepository;

@CrossOrigin(origins = "http://localhost:8080")
@RestController
@RequestMapping("/api/ingredient")
public class IngredientController {
    @Autowired
    IngredientRepository ingredientRepository;

    @SuppressWarnings("null")
    @PostMapping("")
    public ResponseEntity<Ingredient> creaEntity(@RequestBody Ingredient req) {
        try {
            Ingredient _ingredient = ingredientRepository.save(req);
            return new ResponseEntity<>(_ingredient, HttpStatus.CREATED);
        } catch (Exception e) {
            return new ResponseEntity<>(null, HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }
}
