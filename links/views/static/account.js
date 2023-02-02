window.addEventListener("DOMContentLoaded", function () {
    const form = document.getElementById("account-form");
    form.addEventListener("submit", (event) => {
        return event.target.checkValidity();
    });

    const refreshInfo = () => {
        let linkRemovers = document.getElementsByClassName("remove-link");
        Array.prototype.forEach.call(linkRemovers, element => {
            element.addEventListener("click", removeLink);
        });

        let linkUppers = document.getElementsByClassName("move-link-up");
        Array.prototype.forEach.call(linkUppers, element => {
            element.addEventListener("click", upLink);
        });

        let linkDowners = document.getElementsByClassName("move-link-down");
        Array.prototype.forEach.call(linkDowners, element => {
            element.addEventListener("click", downLink);
        });

        let linkTitles = document.getElementsByClassName("link-title");
        let curr = 1;
        Array.prototype.forEach.call(linkTitles, element => {
            element.textContent = `Link #${curr}`;
            curr++;
        });
    };

    const removeLink = (event) => {
        let p = event.target.parentElement.parentElement;
        p.remove();
        refreshInfo();
    };

    const upLink = (event) => {
        let p = event.target.parentElement.parentElement;
        p.parentElement.insertBefore(p, p.previousElementSibling);
        refreshInfo();
    };

    const downLink = (event) => {
        let p = event.target.parentElement.parentElement;
        p.parentElement.insertBefore(p, p.nextElementSibling.nextElementSibling);
        refreshInfo();
    };

    const addLink = () => {
        // Calculate the first available index in a very crude fashion
        // We don't really care about index, we use it to associate info
        // about a single link
        let mlds = document.getElementsByClassName("move-link-down");
        let i = 0;

        Array.prototype.forEach.call(mlds, element => {
            let pi = parseInt(element.getAttribute("data-index"));
            if (pi == i) {
                i = pi + 1;
            }
        });

        let template = `
        <div class="link-entry">
            <div class="link-edit">
              <label class="italic link-title">Link #${i + 1}</label>
    
              <label class="sub-label" for="links_${i}_title">Title</label>
              <input
                type="text"
                name="links_title[]"
                id="links_${i}_title"
                maxlength="128"
                value="New Link!"
                required
              />

              <label class="sub-label" for="links_${i}_url">URL</label>
              <input
                type="url"
                name="links_url[]"
                id="links_${i}_url"
                value="https://links.com"
                required
              />
            </div>
    
            <div class="link-control">
              <button
                class="small-button remove-link"
                type="button"
                data-index="${i}"
              >
                ❌
              </button>
              <button
                class="small-button move-link-up"
                type="button"
                data-index="${i}"
              >
                ⬆
              </button>
              <button
                class="small-button move-link-down"
                type="button"
                data-index="${i}"
              >
                ⬇
              </button>
            </div>
          </div>`;

        let container = document.getElementsByClassName('links-container')[0];
        container.insertAdjacentHTML("beforeend", template);

        refreshInfo();
    };

    let linkAdder = document.getElementsByClassName("add-link")[0];
    linkAdder.addEventListener("click", addLink);

    refreshInfo();
});
