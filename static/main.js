/* ========= 상수·상태 ========= */
const PAGE_SIZE = 12,
  SIDEBAR_MIN = 240,
  SIDEBAR_MAX = 600;
let lastSidebarWidth = 320,
  isSidebarCollapsed = false;

let currentPage = 1,
  selectedTags = [],
  tagSearchKeyword = "",
  tagPage = 1,
  tagTotalPage = 1,
  tagPageSize = 30,
  tagListData = [],
  newsSearchKeyword = "",
  cardItems = [],
  totalCount = 0,
  isLoading = false,
  isEndOfList = false;

/* ========= 사이드바 모바일 ========= */
function openSidebar() {
  const sb = document.getElementById("sidebar");
  sb.classList.remove("closed", "hidden");
  sb.style.width = sb.style.padding = sb.style.overflow = "";
  sb.classList.add(
    "fixed",
    "left-0",
    "top-[65px]",
    "h-[calc(100vh-65px)]",
    "bg-white",
    "w-[80vw]",
    "max-w-[400px]",
    "z-30",
  );
  document.getElementById("sidebarOverlay").classList.remove("hidden");
}
function closeSidebar() {
  const sb = document.getElementById("sidebar");
  sb.classList.add("hidden");
  sb.classList.remove(
    "fixed",
    "left-0",
    "top-[65px]",
    "h-[calc(100vh-65px)]",
    "bg-white",
    "w-[80vw]",
    "max-w-[400px]",
    "z-30",
  );
  document.getElementById("sidebarOverlay").classList.add("hidden");
}

/* ========= 데스크톱 토글 ========= */
function toggleSidebarDesktop() {
  const sb = document.getElementById("sidebar"),
    handle = document.getElementById("sidebarResizeHandle");
  if (isSidebarCollapsed) {
    sb.classList.remove("closed");
    sb.style.borderRight = "1px solid rgb(229 231 235)";
    handle.style.display = "block";
    isSidebarCollapsed = false;
  } else {
    lastSidebarWidth = sb.offsetWidth;
    sb.classList.add("closed");
    sb.style.borderRight = "none";
    handle.style.display = "none";
    isSidebarCollapsed = true;
  }
}

/* ========= 리사이즈 핸들 ========= */
function initSidebarResize() {
  const sb = document.getElementById("sidebar"),
    h = document.getElementById("sidebarResizeHandle");
  if (!h) return;

  let startX = 0,
    startW = 0,
    resizing = false;
  h.addEventListener("mousedown", (e) => {
    if (isSidebarCollapsed) return;
    resizing = true;
    startX = e.clientX;
    startW = sb.offsetWidth;
    document.body.style.userSelect = "none";
  });
  document.addEventListener("mousemove", (e) => {
    if (!resizing) return;
    let w = startW + (e.clientX - startX);
    w = Math.max(SIDEBAR_MIN, Math.min(SIDEBAR_MAX, w));
    sb.style.width = w + "px";
  });
  document.addEventListener("mouseup", () => {
    if (resizing) {
      resizing = false;
      lastSidebarWidth = sb.offsetWidth;
      document.body.style.userSelect = "";
    }
  });
}

function attachTooltip(btn, text) {
  let tipEl = null;
  function show() {
    tipEl = document.createElement("div");
    tipEl.className = "tooltip-box show";
    tipEl.textContent = text;
    document.body.appendChild(tipEl);

    const rect = btn.getBoundingClientRect();
    tipEl.style.top = window.scrollY + rect.top - tipEl.offsetHeight - 6 + "px";
    tipEl.style.left =
      window.scrollX +
      rect.left +
      rect.width / 2 -
      tipEl.offsetWidth / 2 +
      "px";
  }
  function hide() {
    if (tipEl) {
      tipEl.remove();
      tipEl = null;
    }
  }
  btn.addEventListener("mouseenter", show);
  btn.addEventListener("mouseleave", hide);
  btn.addEventListener("mousedown", hide);
}

/* ========= 클릭 후 임시 툴팁 ========= */
function showTempTooltip(btn, message, duration = 1200) {
  const tip = document.createElement("div");
  tip.className = "tooltip-box show";
  tip.textContent = message;
  document.body.appendChild(tip);

  const rect = btn.getBoundingClientRect();
  tip.style.top = window.scrollY + rect.top - tip.offsetHeight - 6 + "px";
  tip.style.left =
    window.scrollX + rect.left + rect.width / 2 - tip.offsetWidth / 2 + "px";

  setTimeout(() => tip.remove(), duration);
}

/* ========= DOMContentLoaded ========= */
document.addEventListener("DOMContentLoaded", () => {
  loadTagList();
  resetAndLoadNews();
  initSidebarResize();

  /* 버튼·이벤트 바인딩 */
  document.getElementById("sidebarOpenBtn").onclick = () => {
    window.innerWidth < 768 ? openSidebar() : toggleSidebarDesktop();
  };
  document.getElementById("sidebarCollapseBtnDesktop").onclick =
    toggleSidebarDesktop;
  document.getElementById("sidebarCloseBtn").onclick = closeSidebar;
  document.getElementById("sidebarOverlay").onclick = closeSidebar;

  document.getElementById("searchBtn").onclick = () => {
    newsSearchKeyword = document.getElementById("searchInput").value.trim();
    resetAndLoadNews();
  };
  document.getElementById("tagSearchBtn").onclick = () => {
    tagSearchKeyword = document.getElementById("tagSearch").value.trim();
    tagPage = 1;
    loadTagList();
  };
  document.getElementById("tagPrevBtn").onclick = () => {
    if (tagPage > 1) {
      tagPage--;
      loadTagList();
    }
  };
  document.getElementById("tagNextBtn").onclick = () => {
    if (tagPage < tagTotalPage) {
      tagPage++;
      loadTagList();
    }
  };
  document.getElementById("searchInput").addEventListener("keyup", (e) => {
    if (e.key === "Enter") {
      newsSearchKeyword = e.target.value.trim();
      resetAndLoadNews();
    }
  });
  document.getElementById("tagSearch").addEventListener("keyup", (e) => {
    if (e.key === "Enter") {
      tagSearchKeyword = e.target.value.trim();
      tagPage = 1;
      loadTagList();
    }
  });

  document
    .getElementById("mainScrollArea")
    .addEventListener("scroll", onMainScroll);

  /* 창 크기 변경 시 정합성 유지 */
  window.addEventListener("resize", () => {
    const sb = document.getElementById("sidebar"),
      handle = document.getElementById("sidebarResizeHandle");

    if (window.innerWidth >= 768) {
      /* 모바일 클래스 제거 */
      closeSidebar();

      /* 접힘 상태 유지 */
      if (isSidebarCollapsed) {
        sb.classList.add("closed");
        sb.style.borderRight = "none";
        handle.style.display = "none";
      } else {
        sb.classList.remove("closed");
        sb.style.borderRight = "1px solid rgb(229 231 235)";
        handle.style.display = "block";
      }
    }
  });
});

/* ========= 태그 API ========= */
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
  const list = document.getElementById("tagList");
  list.innerHTML = "";
  tagListData.forEach((tag) => {
    const checked = selectedTags.some((t) => t.id === tag.id);

    const label = document.createElement("label");
    label.className =
      "flex items-center gap-2 px-3 py-2 rounded-lg border border-transparent hover:bg-blue-50 cursor-pointer" +
      (checked ? " bg-blue-100 border-blue-300 font-semibold" : " bg-white");

    const cb = document.createElement("input");
    cb.type = "checkbox";
    cb.className = "accent-blue-500 mr-2 w-4 h-4";
    cb.checked = checked;
    cb.onchange = () => {
      if (cb.checked) {
        selectedTags.push({ id: tag.id, name: tag.name });
      } else {
        selectedTags = selectedTags.filter((t) => t.id !== tag.id);
      }
      renderTagList();
      renderSelectedTags();
      resetAndLoadNews();
    };

    const chip = document.createElement("span");
    chip.className =
      "inline-flex items-center px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium leading-tight" +
      (checked ? " bg-blue-600 text-white border-blue-500" : "");
    chip.textContent = tag.name;

    const cnt = document.createElement("span");
    cnt.className = "ml-auto text-gray-400 text-xs";
    cnt.textContent = tag.news_count;

    label.append(cb, chip, cnt);
    list.appendChild(label);
  });
  renderSelectedTags();
}
function renderSelectedTags() {
  const wrap = document.getElementById("selectedTags");
  wrap.innerHTML = "";
  selectedTags.forEach((tag) => {
    const span = document.createElement("span");
    span.className =
      "inline-flex items-center px-3 py-0.5 rounded-2xl border text-white bg-blue-600 border-blue-500 text-xs font-medium mr-1 mb-1";
    span.textContent = tag.name;

    const x = document.createElement("span");
    x.className = "cursor-pointer ml-1 text-base font-bold";
    x.textContent = "×";
    x.onclick = () => {
      selectedTags = selectedTags.filter((t) => t.id !== tag.id);
      renderTagList();
      resetAndLoadNews();
    };
    span.appendChild(x);
    wrap.appendChild(span);
  });
}

/* ========= 뉴스 API ========= */
function resetAndLoadNews() {
  currentPage = 1;
  cardItems = [];
  isLoading = false;
  isEndOfList = false;
  document.getElementById("cardList").innerHTML = "";
  document.getElementById("newsCount").textContent = "";
  loadNewsPage();
}
function loadNewsPage() {
  if (isLoading || isEndOfList) return;
  isLoading = true;
  document.getElementById("loading").style.display = "";

  let url = `/api/whatsnews?limit=${PAGE_SIZE}&offset=${cardItems.length}`;
  if (selectedTags.length)
    url += `&tags=${selectedTags.map((t) => t.id).join(",")}`;
  if (newsSearchKeyword)
    url += `&search=${encodeURIComponent(newsSearchKeyword)}`;

  fetch(url)
    .then((r) => r.json())
    .then((data) => {
      const items = data.items || [];
      cardItems = cardItems.concat(items);
      totalCount = data.total || 0;
      appendCards(items);
      document.getElementById("newsCount").textContent = `전체 ${totalCount}건`;
      if (cardItems.length >= totalCount || !items.length) isEndOfList = true;
    })
    .finally(() => {
      isLoading = false;
      document.getElementById("loading").style.display = "none";
    });
}

/* ========= 카드 렌더 ========= */
function appendCards(items) {
  const list = document.getElementById("cardList");
  items.forEach((it) => {
    const card = document.createElement("div");
    card.className =
      "news-card bg-white border border-blue-100 rounded-2xl shadow-lg flex flex-col px-6 py-6 min-h-[200px]";

    /* ─ Title ─ */
    const title = document.createElement("div");
    title.className = "text-lg font-semibold mb-3 leading-tight break-all";
    if (it.source_url) {
      let u = it.source_url;
      if (!/^https?:\/\//.test(u))
        u = "https://aws.amazon.com" + (u.startsWith("/") ? u : "/" + u);

      /* ─ 링크 ─ */
      const a = document.createElement("a");
      a.href = u;
      a.target = "_blank";
      a.textContent = it.title;
      a.className = "text-blue-700 hover:underline";
      title.appendChild(a);

      /* ─ Copy 버튼 (SVG 아이콘) ─ */
      const copyBtn = document.createElement("button");
      copyBtn.type = "button";
      copyBtn.className =
        "ml-2 p-1 text-gray-400 hover:text-blue-600 focus:outline-none cursor-pointer";

      /* 아이콘 정의 */
      const iconDefault = `
        <svg xmlns="http://www.w3.org/2000/svg"
             viewBox="0 0 115.77 122.88"
             class="w-4 h-4 fill-current"><path d="M89.62,13.96v7.73h12.19h0.01v0.02c3.85,0.01,7.34,1.57,9.86,4.1c2.5,2.51,4.06,5.98,4.07,9.82h0.02v0.02 v73.27v0.01h-0.02c-0.01,3.84-1.57,7.33-4.1,9.86c-2.51,2.5-5.98,4.06-9.82,4.07v0.02h-0.02h-61.7H40.1v-0.02 c-3.84-0.01-7.34-1.57-9.86-4.1c-2.5-2.51-4.06-5.98-4.07-9.82h-0.02v-0.02V92.51H13.96h-0.01v-0.02c-3.84-0.01-7.34-1.57-9.86-4.1 c-2.5-2.51-4.06-5.98-4.07-9.82H0v-0.02V13.96v-0.01h0.02c0.01-3.85,1.58-7.34,4.1-9.86c2.51-2.5,5.98-4.06,9.82-4.07V0h0.02h61.7 h0.01v0.02c3.85,0.01,7.34,1.57,9.86,4.1c2.5,2.51,4.06,5.98,4.07,9.82h0.02V13.96L89.62,13.96z M79.04,21.69v-7.73v-0.02h0.02 c0-0.91-0.39-1.75-1.01-2.37c-0.61-0.61-1.46-1-2.37-1v0.02h-0.01h-61.7h-0.02v-0.02c-0.91,0-1.75,0.39-2.37,1.01 c-0.61,0.61-1,1.46-1,2.37h0.02v0.01v64.59v0.02h-0.02c0,0.91,0.39,1.75,1.01,2.37c0.61,0.61,1.46,1,2.37,1v-0.02h0.01h12.19V35.65 v-0.01h0.02c0.01-3.85,1.58-7.34,4.1-9.86c2.51-2.5,5.98-4.06,9.82-4.07v-0.02h0.02H79.04L79.04,21.69z M105.18,108.92V35.65v-0.02 h0.02c0-0.91-0.39-1.75-1.01-2.37c-0.61-0.61-1.46-1-2.37-1v0.02h-0.01h-61.7h-0.02v-0.02c-0.91,0-1.75,0.39-2.37,1.01 c-0.61,0.61-1,1.46-1,2.37h0.02v0.01v73.27v0.02h-0.02c0,0.91,0.39,1.75,1.01,2.37c0.61,0.61,1.46,1,2.37,1v-0.02h0.01h61.7h0.02 v0.02c0.91,0,1.75-0.39,2.37-1.01c0.61-0.61,1-1.46,1-2.37h-0.02V108.92L105.18,108.92z"/></svg>`;
      const iconChecked = `
        <svg xmlns="http://www.w3.org/2000/svg"
             viewBox="0 0 128 128"
             class="w-4 h-4 fill-current"><path d="M0 52.88l22.68-0.3c8.76 5.05 16.6 11.59 23.35 19.86C63.49 43.49 83.55 19.77 105.6 0h17.28C92.05 34.25 66.89 70.92 46.77 109.76C36.01 86.69 20.96 67.27 0 52.88Z"/></svg>`;

      copyBtn.innerHTML = iconDefault;
      attachTooltip(copyBtn, "링크 복사");

      copyBtn.onclick = (e) => {
        e.preventDefault();
        navigator.clipboard
          .writeText(u)
          .then(() => {
            copyBtn.classList.replace("text-gray-400", "text-green-600");
            copyBtn.classList.remove("hover:text-blue-600");
            copyBtn.innerHTML = iconChecked;

            setTimeout(() => {
              copyBtn.classList.replace("text-green-600", "text-gray-400");
              copyBtn.classList.add("hover:text-blue-600");
              copyBtn.innerHTML = iconDefault;
            }, 1200);

            /* ▷ 복사 완료 툴팁 */
            showTempTooltip(copyBtn, "복사 완료");
          })
          .catch(() => alert("클립보드 복사 실패"));
      };
      title.appendChild(copyBtn);
    } else {
      title.textContent = it.title;
    }

    /* ─ Date ─ */
    const date = document.createElement("div");
    date.className = "text-gray-500 text-sm mb-2";
    date.textContent = it.source_created_at
      ? it.source_created_at.slice(0, 10)
      : "";

    /* ─ Tags ─ */
    const tagWrap = document.createElement("div");
    tagWrap.className = "flex flex-wrap gap-2 mb-3";
    (it.tags || []).forEach((tg) => {
      const s = document.createElement("span");
      s.className =
        "inline-flex items-center px-3 py-0.5 rounded-2xl border text-blue-700 bg-blue-50 border-blue-200 text-xs font-medium leading-tight";
      s.textContent = tg.name;
      tagWrap.appendChild(s);
    });

    /* ─ Body ─ */
    const body = document.createElement("div");
    body.className = "text-base text-gray-900 mt-1";
    if (it.content) body.innerHTML = it.content;

    card.append(title, date, tagWrap, body);
    list.appendChild(card);
  });
}

/* ========= 무한 스크롤 ========= */
function onMainScroll(e) {
  const el = e.target;
  if (isLoading || isEndOfList) return;
  if (el.scrollHeight - el.scrollTop - el.clientHeight < 80) loadNewsPage();
}
