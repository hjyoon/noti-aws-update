const PAGE_SIZE = 12;
let currentPage = 1;
let selectedTags = [];
let tagSearchKeyword = "";
let tagPage = 1;
let tagTotalPage = 1;
let tagPageSize = 30;
let tagListData = [];
let newsSearchKeyword = "";
let cardItems = [];
let totalCount = 0;
let totalPage = 1;
let isLoading = false;
let isEndOfList = false;

// 사이드바 토글 로직
function openSidebar() {
  document.getElementById("sidebar").classList.remove("hidden");
  document
    .getElementById("sidebar")
    .classList.add(
      "fixed",
      "left-0",
      "top-0",
      "h-screen",
      "bg-white",
      "w-[80vw]",
      "max-w-[400px]",
      "z-30",
    );
  document.getElementById("sidebarOverlay").classList.remove("hidden");
}
function closeSidebar() {
  document.getElementById("sidebar").classList.add("hidden");
  document
    .getElementById("sidebar")
    .classList.remove(
      "fixed",
      "left-0",
      "top-0",
      "h-screen",
      "bg-white",
      "w-[80vw]",
      "max-w-[400px]",
      "z-30",
    );
  document.getElementById("sidebarOverlay").classList.add("hidden");
}

document.addEventListener("DOMContentLoaded", () => {
  loadTagList();
  resetAndLoadNews();

  // 햄버거 버튼/사이드바 오픈
  document.getElementById("sidebarOpenBtn").onclick = openSidebar;
  document.getElementById("sidebarCloseBtn").onclick = closeSidebar;
  document.getElementById("sidebarOverlay").onclick = closeSidebar;

  document.getElementById("searchBtn").onclick = function () {
    newsSearchKeyword = document.getElementById("searchInput").value.trim();
    resetAndLoadNews();
  };
  document.getElementById("tagSearchBtn").onclick = function () {
    tagSearchKeyword = document.getElementById("tagSearch").value.trim();
    tagPage = 1;
    loadTagList();
  };
  document.getElementById("tagPrevBtn").onclick = function () {
    if (tagPage > 1) {
      tagPage--;
      loadTagList();
    }
  };
  document.getElementById("tagNextBtn").onclick = function () {
    if (tagPage < tagTotalPage) {
      tagPage++;
      loadTagList();
    }
  };
  document
    .getElementById("searchInput")
    .addEventListener("keyup", function (event) {
      if (event.key === "Enter") {
        newsSearchKeyword = this.value.trim();
        resetAndLoadNews();
      }
    });
  document
    .getElementById("tagSearch")
    .addEventListener("keyup", function (event) {
      if (event.key === "Enter") {
        tagSearchKeyword = this.value.trim();
        tagPage = 1;
        loadTagList();
      }
    });
  document
    .getElementById("mainScrollArea")
    .addEventListener("scroll", onMainScroll);

  // 모바일에서 사이드바 resize시 자동 close
  window.addEventListener("resize", () => {
    if (window.innerWidth >= 768) closeSidebar();
  });
});

function resetAndLoadNews() {
  currentPage = 1;
  cardItems = [];
  isLoading = false;
  isEndOfList = false;
  document.getElementById("cardList").innerHTML = "";
  document.getElementById("newsCount").textContent = "";
  loadNewsPage();
}
function loadTagList() {
  let url = `/api/tags?limit=${tagPageSize}&offset=${(tagPage - 1) * tagPageSize}`;
  if (tagSearchKeyword) url += `&name=${encodeURIComponent(tagSearchKeyword)}`;
  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      tagListData = data.items;
      tagTotalPage = data.total_page || 1;
      renderTagList();
      document.getElementById("tagPageInfo").textContent =
        `Page ${data.page} / ${tagTotalPage}`;
    });
}
function renderTagList() {
  const tagListDiv = document.getElementById("tagList");
  tagListDiv.innerHTML = "";
  tagListData.forEach((tag) => {
    let label = document.createElement("label");
    label.className =
      "flex items-center gap-2 px-3 py-2 rounded-lg border border-transparent hover:bg-blue-50 cursor-pointer" +
      (selectedTags.some((t) => t.id === tag.id)
        ? " bg-blue-100 border-blue-300"
        : " bg-white");
    if (selectedTags.some((t) => t.id === tag.id))
      label.classList.add("font-semibold");
    let cb = document.createElement("input");
    cb.type = "checkbox";
    cb.className = "tag-checkbox accent-blue-500 mr-2 w-4 h-4";
    cb.value = tag.id;
    cb.checked = selectedTags.some((t) => t.id === tag.id);
    cb.onchange = function () {
      if (this.checked) {
        if (!selectedTags.some((t) => t.id === tag.id)) {
          selectedTags.push({ id: tag.id, name: tag.name });
        }
      } else {
        selectedTags = selectedTags.filter((t) => t.id !== tag.id);
      }
      renderTagList();
      renderSelectedTags();
      resetAndLoadNews();
    };
    let tagChip = document.createElement("span");
    tagChip.className =
      "inline-block px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium";
    if (selectedTags.some((t) => t.id === tag.id)) {
      tagChip.className += " bg-blue-600 text-white border-blue-500";
    }
    tagChip.textContent = tag.name;
    let countSpan = document.createElement("span");
    countSpan.className = "ml-auto text-gray-400 text-xs";
    countSpan.textContent = tag.news_count;
    label.appendChild(cb);
    label.appendChild(tagChip);
    label.appendChild(countSpan);
    tagListDiv.appendChild(label);
  });
  renderSelectedTags();
}
function renderSelectedTags() {
  const selDiv = document.getElementById("selectedTags");
  selDiv.innerHTML = "";
  selectedTags.forEach((tag) => {
    let span = document.createElement("span");
    span.className =
      "inline-flex items-center px-3 py-0.5 rounded-2xl border text-white bg-blue-600 border-blue-500 text-xs font-medium mr-1 mb-1";
    span.textContent = tag.name;
    let x = document.createElement("span");
    x.className = "tag-remove cursor-pointer ml-1 text-base font-bold";
    x.textContent = "×";
    x.title = "태그 해제";
    x.onclick = function () {
      selectedTags = selectedTags.filter((t) => t.id !== tag.id);
      renderTagList();
      resetAndLoadNews();
    };
    span.appendChild(x);
    selDiv.appendChild(span);
  });
}
function loadNewsPage() {
  if (isLoading || isEndOfList) return;
  isLoading = true;
  document.getElementById("loading").style.display = "";

  let url = `/api/whatsnews?limit=${PAGE_SIZE}&offset=${cardItems.length}`;
  if (selectedTags.length > 0)
    url += `&tags=${selectedTags.map((t) => t.id).join(",")}`;
  if (newsSearchKeyword)
    url += `&search=${encodeURIComponent(newsSearchKeyword)}`;

  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      let items = data.items || [];
      cardItems = cardItems.concat(items);
      totalCount = data.total || 0;
      totalPage = data.total_page || 1;

      fillCardList(cardItems);

      document.getElementById("newsCount").textContent = `전체 ${totalCount}건`;

      if (cardItems.length >= totalCount || items.length === 0) {
        isEndOfList = true;
      }
    })
    .finally(() => {
      isLoading = false;
      document.getElementById("loading").style.display = "none";
    });
}
function fillCardList(items) {
  const cardList = document.getElementById("cardList");
  cardList.innerHTML = "";
  items.forEach((item) => {
    let card = document.createElement("div");
    card.className =
      "card bg-white border border-blue-100 rounded-2xl shadow-lg flex flex-col min-w-0 min-h-[200px] max-w-3xl mx-auto px-6 py-6";
    let titleDiv = document.createElement("div");
    titleDiv.className =
      "card-title text-lg font-semibold mb-3 leading-tight break-all";
    if (item.source_url) {
      let url = item.source_url;
      if (!/^https?:\/\//.test(url)) {
        url =
          "https://aws.amazon.com" + (url.startsWith("/") ? url : "/" + url);
      }
      let a = document.createElement("a");
      a.href = url;
      a.textContent = item.title;
      a.target = "_blank";
      a.className = "text-blue-700 hover:underline";
      titleDiv.appendChild(a);
    } else {
      titleDiv.textContent = item.title;
    }
    card.appendChild(titleDiv);

    let dateDiv = document.createElement("div");
    dateDiv.className = "card-date text-gray-500 text-sm mb-2";
    dateDiv.textContent = item.source_created_at
      ? item.source_created_at.substring(0, 10)
      : "";
    card.appendChild(dateDiv);

    let tagsDiv = document.createElement("div");
    tagsDiv.className = "card-tags flex flex-wrap gap-2 mb-3 min-h-[1.8em]";
    if (item.tags && item.tags.length) {
      item.tags.forEach((tag) => {
        let span = document.createElement("span");
        span.className =
          "inline-block px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium";
        span.textContent = tag.name;
        tagsDiv.appendChild(span);
      });
    }
    card.appendChild(tagsDiv);

    let contentDiv = document.createElement("div");
    contentDiv.className = "card-content text-base text-gray-900 mt-1";
    if (item.content) {
      contentDiv.innerHTML = item.content;
    }
    card.appendChild(contentDiv);

    cardList.appendChild(card);
  });
}
function onMainScroll(e) {
  const el = e.target;
  if (isLoading || isEndOfList) return;
  if (el.scrollHeight - el.scrollTop - el.clientHeight < 80) {
    loadNewsPage();
  }
}
