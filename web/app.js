async function loadDashboard() {

    const response = await fetch("/api/dashboard");
    const data = await response.json();

    document.getElementById("available-skis").textContent =
        data.availableSkis;

    document.getElementById("rented-skis").textContent =
        data.rentedSkis;

    document.getElementById("total-skis").textContent =
        data.totalSkis;

    document.getElementById("registered-persons").textContent =
        data.registeredPersons;

}

loadDashboard();


async function loadSkis() {
    const response = await fetch("/api/skis");

    if (!response.ok) {
        throw new Error("Could not load skis");
    }

    const skis = await response.json();
    const list = document.getElementById("ski-list");

    list.innerHTML = "";

    for (const ski of skis) {
        const item = document.createElement("li");

        item.textContent = ski.ID + " - " + ski.Model +
            " - " + (ski.Available ? "verfügbar" : "ausgeliehen");

        list.appendChild(item);
    }
}

document.getElementById("show-skis").addEventListener("click", async function() {
    document.getElementById("ski-management").hidden = false;

    try {
        await loadSkis();
    } catch (error) {
        document.getElementById("ski-feedback").textContent =
            "Fehler beim Laden der Ski.";
    }
});

document.getElementById("ski-form").addEventListener("submit", async function(event) {
    event.preventDefault();

    const input = document.getElementById("ski-model");
    const feedback = document.getElementById("ski-feedback");

    try {
        const response = await fetch("/api/skis", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                model: input.value
            })
        });

        if (!response.ok) {
            throw new Error("Could not add ski");
        }

        input.value = "";
        feedback.textContent = "Ski erfolgreich hinzugefügt!";

        await loadSkis();
        await loadDashboard();
    } catch (error) {
        feedback.textContent = "Fehler beim Hinzufügen.";
    }
});



async function loadPersons() {
    const response = await fetch("/api/persons");

    if (!response.ok) {
        throw new Error("Could not load persons");
    }

    const persons = await response.json();
    const list = document.getElementById("person-list");

    list.innerHTML = "";

    for (const person of persons) {
        const item = document.createElement("li");

        item.textContent = person.ID + " - " +
            person.FirstName + " " + person.LastName +
            " - " + person.DocumentType;

        list.appendChild(item);
    }
}

document.getElementById("show-persons").addEventListener("click", async function() {
    document.getElementById("ski-management").hidden = true;
    document.getElementById("person-management").hidden = false;

    try {
        await loadPersons();
    } catch (error) {
        document.getElementById("person-feedback").textContent =
            "Fehler beim Laden der Personen.";
    }
});

document.getElementById("show-skis").addEventListener("click", function() {
    document.getElementById("person-management").hidden = true;
});

document.getElementById("person-form").addEventListener("submit", async function(event) {
    event.preventDefault();

    const feedback = document.getElementById("person-feedback");

    const person = {
        firstName: document.getElementById("first-name").value,
        lastName: document.getElementById("last-name").value,
        documentType: document.getElementById("document-type").value,
        documentNr: document.getElementById("document-nr").value
    };

    try {
        const response = await fetch("/api/persons", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(person)
        });

        if (!response.ok) {
            throw new Error("Could not add person");
        }

        document.getElementById("person-form").reset();

        feedback.textContent = "Person erfolgreich hinzugefügt!";

        await loadPersons();
        await loadDashboard();
        await loadDocumentPersons();
    } catch (error) {
        feedback.textContent = "Fehler beim Hinzufügen.";
    }
});






async function loadRentalOptions() {
    const dateInput = document.getElementById("rental-date");
    const durationInput = document.getElementById("rental-duration");
    const unitInput = document.getElementById("rental-unit");

    if (!dateInput.value) {
        const date = new Date(Date.now() + 10 * 60000);
        const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);

        dateInput.value = local.toISOString().slice(0, 16);
    }

    const query = new URLSearchParams({
        startDate: dateInput.value,
        quantity: durationInput.value,
        unit: unitInput.value
    });

    const [personsResponse, skisResponse] = await Promise.all([
        fetch("/api/persons"),
        fetch("/api/skis/available?" + query.toString())
    ]);

    if (!personsResponse.ok || !skisResponse.ok) {
        throw new Error("Could not load rental options");
    }

    const persons = await personsResponse.json();
    const skis = await skisResponse.json();

    const personSelect = document.getElementById("rental-person");
    const skiSelect = document.getElementById("rental-ski");

    personSelect.innerHTML = "";
    skiSelect.innerHTML = "";

    for (const person of persons) {
        const option = document.createElement("option");
        option.value = person.ID;
        option.textContent = person.FirstName + " " + person.LastName;
        personSelect.appendChild(option);
    }

    for (const ski of skis) {
        const option = document.createElement("option");
        option.value = ski.ID;
        option.textContent = ski.Model + " (" + ski.ID + ")";
        skiSelect.appendChild(option);
    }

    document.querySelector("#rental-form button").disabled =
        persons.length === 0 || skis.length === 0;

    await updateRentalPrice();
}





            document.getElementById("rental-date").addEventListener("change", async function() {
                try {
                    await loadRentalOptions();
                } catch (error) {
                    document.getElementById("rental-feedback").textContent =
                        "Fehler beim Laden der verfügbaren Ski.";
                }
            });

            document.getElementById("rental-duration").addEventListener("change", async function() {
                try {
                    await loadRentalOptions();
                } catch (error) {
                    document.getElementById("rental-feedback").textContent =
                        "Fehler beim Laden der verfügbaren Ski.";
                }
            });





document.getElementById("show-rentals").addEventListener("click", async function() {
    document.getElementById("ski-management").hidden = true;
    document.getElementById("person-management").hidden = true;
    document.getElementById("rental-management").hidden = false;

    try {
        await loadRentalOptions();
        await loadRentals();
    } catch (error) {
        document.getElementById("rental-feedback").textContent =
            "Fehler beim Laden der Daten.";
    }
});

document.getElementById("show-skis").addEventListener("click", function() {
    document.getElementById("rental-management").hidden = true;
});

document.getElementById("show-persons").addEventListener("click", function() {
    document.getElementById("rental-management").hidden = true;
});

document.getElementById("rental-form").addEventListener("submit", async function(event) {
    event.preventDefault();

    const feedback = document.getElementById("rental-feedback");

    
    const rental = {
        personId: document.getElementById("rental-person").value,
        skiId: document.getElementById("rental-ski").value,
        startDate: document.getElementById("rental-date").value,
        quantity: Number(document.getElementById("rental-duration").value),
        unit: document.getElementById("rental-unit").value
    };


    try {
        const response = await fetch("/api/rentals", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(rental)
        });

        if (!response.ok) {
            throw new Error("Rental failed");
        }

        const savedRental = await response.json();

        feedback.textContent =
            "Ausleihe erfolgreich erstellt! Gesamtpreis: " +
            formatEuro(savedRental.TotalCents);

        await loadRentalOptions();
        await loadDashboard();
        await loadRentals();
    } catch (error) {
        feedback.textContent = "Fehler beim Erstellen der Ausleihe.";
    }
});




async function loadRentals() {
    const response = await fetch("/api/rentals");

    if (!response.ok) {
        throw new Error("Could not load rentals");
    }

    const rentals = await response.json();
    const list = document.getElementById("rental-list");

    list.innerHTML = "";

    for (const rental of rentals) {
        const item = document.createElement("li");

        const text = document.createElement("span");

        
        text.textContent =
            rental.ID +
            " - Person: " + rental.PersonID +
            " - Ski: " + rental.SkiID +
            " - Beginn: " + rental.StartDate.slice(0, 16).replace("T", " ") +
            " - Dauer: " + (rental.Quantity || rental.DurationDays) +
            " " + (rental.Unit === "hours" ? "Stunden" : "Tage") +
            " - Preis: " + formatEuro(rental.TotalCents) +
            " - " + (rental.Returned ? "zurückgegeben" : "aktiv");


        item.appendChild(text);

        if (!rental.Returned) {
            const button = document.createElement("button");

            button.textContent = "Ski zurückgeben";
            button.style.marginLeft = "15px";

            button.addEventListener("click", async function() {
                button.disabled = true;

                try {
                    const response = await fetch("/api/rentals/return", {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json"
                        },
                        body: JSON.stringify({
                            rentalId: rental.ID
                        })
                    });

                    if (!response.ok) {
                        throw new Error("Return failed");
                    }

                    document.getElementById("rental-feedback").textContent =
                        "Ski erfolgreich zurückgegeben!";

                    await loadRentals();
                    await loadDashboard();
                    await loadRentalOptions();
                    await loadSkis();

                } catch (error) {
                    document.getElementById("rental-feedback").textContent =
                        "Fehler bei der Rückgabe.";

                    button.disabled = false;
                }
            });

            item.appendChild(button);
        }

        list.appendChild(item);
    }
}




document.getElementById("nav-dashboard").addEventListener("click", function() {
    document.getElementById("ski-management").hidden = true;
    document.getElementById("person-management").hidden = true;
    document.getElementById("rental-management").hidden = true;
});

document.getElementById("nav-skis").addEventListener("click", function() {
    document.getElementById("show-skis").click();
});

document.getElementById("nav-persons").addEventListener("click", function() {
    document.getElementById("show-persons").click();
});

document.getElementById("nav-rentals").addEventListener("click", function() {
    document.getElementById("show-rentals").click();
});





document.getElementById("nav-dashboard").addEventListener("click", function() {
    document.getElementById("ski-management").hidden = true;
    document.getElementById("person-management").hidden = true;
    document.getElementById("rental-management").hidden = true;

    loadDashboard();
});

document.getElementById("nav-skis").addEventListener("click", function() {
    document.getElementById("show-skis").click();
});

document.getElementById("nav-persons").addEventListener("click", function() {
    document.getElementById("show-persons").click();
});

document.getElementById("nav-rentals").addEventListener("click", function() {
    document.getElementById("show-rentals").click();
});


// Document upload

async function loadDocumentPersons() {
    const response = await fetch("/api/persons");

    if (!response.ok) {
        throw new Error("Could not load persons");
    }

    const persons = await response.json();
    const select = document.getElementById("document-person");

    select.innerHTML = "";

    for (const person of persons) {
        const option = document.createElement("option");

        option.value = person.ID;
        option.textContent = person.FirstName + " " + person.LastName;

        select.appendChild(option);
    }
}

document.getElementById("show-persons").addEventListener("click", async function() {
    try {
        await loadDocumentPersons();
    } catch (error) {
        document.getElementById("document-feedback").textContent =
            "Fehler beim Laden der Personen.";
    }
});

document.getElementById("document-form").addEventListener("submit", async function(event) {
    event.preventDefault();

    const feedback = document.getElementById("document-feedback");
    const file = document.getElementById("document-file").files[0];

    if (!file) {
        feedback.textContent = "Bitte eine Datei auswählen.";
        return;
    }

    if (file.size > 2 * 1024 * 1024) {
        feedback.textContent = "Die Datei darf maximal 2 MB groß sein.";
        return;
    }

    const formData = new FormData();

    formData.append(
        "personId",
        document.getElementById("document-person").value
    );

    formData.append("document", file);

    try {
        const response = await fetch("/api/persons/document", {
            method: "POST",
            body: formData
        });

        if (!response.ok) {
            throw new Error("Upload failed");
        }

        feedback.textContent = "Ausweis erfolgreich hochgeladen!";

        document.getElementById("document-file").value = "";

    } catch (error) {
        feedback.textContent = "Fehler beim Hochladen.";
    }
});




let skiPrices = [];

async function loadPricing() {
    const response = await fetch("/api/prices");

    if (!response.ok) {
        throw new Error("Could not load prices");
    }

    skiPrices = await response.json();

    const select = document.getElementById("price-ski");
    select.innerHTML = "";

    for (const ski of skiPrices) {
        const option = document.createElement("option");

        option.value = ski.skiId;
        option.textContent = ski.model + " (" + ski.skiId + ")";

        select.appendChild(option);
    }

    updatePriceInputs();
}

function updatePriceInputs() {
    const skiID = document.getElementById("price-ski").value;
    const ski = skiPrices.find(item => item.skiId === skiID);

    if (!ski) {
        return;
    }

    document.getElementById("price-hour").value =
        (ski.hourlyCents / 100).toFixed(2);

    document.getElementById("price-day").value =
        (ski.dailyCents / 100).toFixed(2);
}

document.getElementById("price-ski").addEventListener("change", updatePriceInputs);

document.getElementById("show-skis").addEventListener("click", async function() {
    try {
        await loadPricing();
    } catch (error) {
        document.getElementById("price-feedback").textContent =
            "Fehler beim Laden der Preise.";
    }
});


    document.getElementById("price-form").addEventListener("submit", async function(event) {
        event.preventDefault();

        const feedback = document.getElementById("price-feedback");

        const data = {
            skiId: document.getElementById("price-ski").value,
            hourlyCents: Math.round(
                Number(document.getElementById("price-hour").value) * 100
            ),
            dailyCents: Math.round(
                Number(document.getElementById("price-day").value) * 100
            )
        };

        try {
            const response = await fetch("/api/prices", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(data)
            });

            if (!response.ok) {
                throw new Error("Could not save prices");
            }

            feedback.textContent = "Preise erfolgreich gespeichert!";

            await loadPricing();

        } catch (error) {
            feedback.textContent = "Fehler beim Speichern der Preise.";
        }
    });




    
function formatEuro(cents) {
    return (cents / 100).toFixed(2).replace(".", ",") + " €";
}

async function updateRentalPrice() {
    const priceElement = document.getElementById("rental-price");
    const skiID = document.getElementById("rental-ski").value;
    const quantity = Number(document.getElementById("rental-duration").value);
    const unit = document.getElementById("rental-unit").value;

    if (!skiID || quantity < 1) {
        priceElement.textContent = "Gesamtpreis: —";
        return;
    }

    const query = new URLSearchParams({
        skiId: skiID,
        unit: unit,
        quantity: String(quantity)
    });

    try {
        const response = await fetch("/api/quote?" + query.toString());

        if (!response.ok) {
            throw new Error("Price calculation failed");
        }

        const data = await response.json();

        priceElement.textContent =
            "Gesamtpreis: " + formatEuro(data.totalCents);

    } catch (error) {
        priceElement.textContent = "Preis konnte nicht berechnet werden.";
    }
}





document.getElementById("rental-unit").addEventListener("change", async function() {
    try {
        await loadRentalOptions();
    } catch (error) {
        document.getElementById("rental-price").textContent =
            "Preis konnte nicht berechnet werden.";
    }
});

document.getElementById("rental-ski").addEventListener("change", updateRentalPrice);
